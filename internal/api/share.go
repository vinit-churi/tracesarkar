package api

import (
	"bytes"
	"context"
	"image"
	// The card is drawn on the redacted JPEG. Registering the decoder here
	// rather than relying on another package importing it keeps this file
	// honest about what it can read.
	_ "image/jpeg"
	"net/http"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/issue"
	"github.com/vinit-churi/tracesarkar/internal/publish"
	"github.com/vinit-churi/tracesarkar/internal/share"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Handles is the registry of verified accounts a generated post may tag.
type Handles interface {
	Handles(ctx context.Context, authority, ward string) ([]share.Handle, error)
}

// handleShareKit writes the post, in every language and for every channel.
//
// The friction that kills sharing is writing the post, so the platform writes
// it (share kit §1). What makes that safe rather than reckless is that it is
// composed from store.PublicIssueView and nothing else — the same coarsened,
// reference-free view the permalink is built from — so a fact the public page
// may not show cannot reach a post that will be forwarded past it.
func (s *Server) handleShareKit(w http.ResponseWriter, r *http.Request) {
	v, ok := s.lookUpPublicIssue(w, r)
	if !ok {
		return
	}

	f, unavailable := s.factsFor(r, v)
	kit, err := share.ComposeKit(f)
	if err != nil {
		// A post that fails the tone check is a bug here, not a reason to show
		// the user something unchecked.
		s.log.Error("could not compose a share kit", "issue_id", v.IssueID, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not write the post for this issue")
		return
	}

	body := map[string]any{
		"issue_id":   kit.IssueID,
		"posts":      kit.Posts,
		"card":       kit.Card.Lines(),
		"card_image": "/v1/public/issue/" + v.IssueID + "/card",
		"note": "Every fact in these posts is on the issue page, with the " +
			"document it came from. Nothing here is sent for you.",
	}
	if unavailable {
		// Said out loud rather than left as an absence: an untagged post reads
		// exactly like a finished one, and a client that cannot tell will show
		// it as finished.
		body["handles_unavailable"] = true
	}
	if d := draftLanguages(kit); len(d) > 0 {
		body["draft_languages"] = d
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, body)
}

// draftLanguages names the languages no native speaker has read yet, so the
// client can mark them and the user can edit before sending.
func draftLanguages(k share.Kit) []string {
	seen := map[string]bool{}
	var out []string
	for _, lang := range share.Languages() {
		for _, ch := range share.Channels() {
			p, ok := k.Posts[string(lang)+"."+string(ch)]
			if ok && p.Review == share.Draft && !seen[string(lang)] {
				seen[string(lang)] = true
				out = append(out, string(lang))
			}
		}
	}
	return out
}

// factsFor narrows the public view into what a post may state.
//
// It reports separately whether the registry could be read, because "no
// handle" and "we could not look" are different answers and only one of them
// is safe to render silently.
func (s *Server) factsFor(r *http.Request, v store.PublicIssueView) (share.Facts, bool) {
	d := publish.Derive(publish.Capture{
		ReportID: v.IssueID, Lat: v.Lat, Lon: v.Lon,
		Category: v.Category, Subcategory: v.Subcategory,
	}, v.Redaction)

	f := share.Facts{
		IssueID: v.IssueID,
		// A relative link, made absolute against the host serving this
		// request: the permalink must be the page the reader can open, and
		// hard-coding a domain here is how a staging link ends up forwarded.
		Permalink:        permalink(r, v.IssueID),
		Authority:        v.Authority,
		Ward:             v.Ward,
		Category:         d.Category,
		Subcategory:      d.Subcategory,
		FirstSeen:        v.FirstSeen,
		Sightings:        v.Sightings,
		Status:           v.Status,
		Lodged:           v.NotifiedAt != nil,
		LodgedOn:         v.NotifiedAt,
		DeadlineHours:    v.DeadlineHours,
		DeadlineShortRef: v.DeadlineShortRef,
		DeadlineCitation: v.DeadlineCitation,
		DeadlineState:    string(clockState(v)),
	}
	// Deliberately absent: Lat, Lon and Reference. The position is a cell on
	// the page and nothing on a forwarded post, and the reference belongs to
	// the person who lodged the complaint.

	if s.handles == nil {
		return f, false
	}
	hs, err := s.handles.Handles(r.Context(), v.Authority, v.Ward)
	if err != nil {
		s.log.Error("could not read the handle registry", "error", err.Error())
		return f, true
	}
	f.Handles = hs
	return f, false
}

// handleShareCard draws the fact block under the redacted photograph.
//
// Every read goes through publish.Derive first, exactly as the image route
// does: a photograph the scan held is a 404 here too, however much more useful
// an annotated version of it would be.
func (s *Server) handleShareCard(w http.ResponseWriter, r *http.Request) {
	if s.blobs == nil {
		writeError(w, http.StatusNotImplemented, "media is not configured on this server")
		return
	}
	v, ok := s.lookUpPublicIssue(w, r)
	if !ok {
		return
	}

	d := publish.Derive(publish.Capture{ReportID: v.IssueID, Lat: v.Lat, Lon: v.Lon}, v.Redaction)
	if d.ImageKey == "" {
		writeError(w, http.StatusNotFound, "no published photograph for this issue")
		return
	}
	stored, err := s.blobs.Get(r.Context(), d.ImageKey)
	if err != nil {
		s.log.Error("could not read a redacted copy", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that photograph")
		return
	}
	// Sized for the web first, then annotated: the fact block is drawn at the
	// size it will be read at, so the text is never resampled.
	sized, err := publish.ForWeb(stored)
	if err != nil {
		s.log.Error("could not size a photograph for the card", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that photograph")
		return
	}
	photo, _, err := image.Decode(bytes.NewReader(sized))
	if err != nil {
		s.log.Error("could not decode a redacted copy", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that photograph")
		return
	}

	f, _ := s.factsFor(r, v)
	body, err := share.Encode(photo, share.CardFor(f))
	if err != nil {
		s.log.Error("could not draw a share card", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not draw the card for this issue")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	// Public, because it is drawn on the redacted copy and carries no fact the
	// page does not carry.
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// clockState is the deadline as the page reads it, so a post and the page it
// links to can never disagree about whether time has run out.
func clockState(v store.PublicIssueView) issue.ClockState {
	if v.DeadlineHours <= 0 || v.DeadlineCitation == "" {
		return issue.NoDeadline
	}
	return issue.NewClock(v.NotifiedAt, v.DeadlineHours).At(time.Now().UTC()).State
}

func permalink(r *http.Request, issueID string) string {
	scheme := "https"
	// Behind Cloud Run the TLS terminates upstream, so r.TLS is nil on a
	// request that reached the user over HTTPS. The forwarded header is the
	// only honest signal, and http is the fallback only for a local run.
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") == "" &&
		(strings.HasPrefix(r.Host, "localhost") || strings.HasPrefix(r.Host, "127.0.0.1")) {
		scheme = "http"
	}
	return scheme + "://" + r.Host + "/issue/" + issueID
}
