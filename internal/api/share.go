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
	"sync"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/issue"
	"github.com/vinit-churi/tracesarkar/internal/polish"
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

	// The model, if one is configured. It never sees the fact set and never
	// writes a post: it rewords the sentences the platform already composed,
	// and anything it returns that cannot be proved to say the same facts is
	// thrown away in favour of the composed text (ADR 0021).
	var polishMeta map[string]any
	if s.polish != nil {
		polishMeta = s.polishKit(r.Context(), f, &kit, v.IssueID)
	}

	body := map[string]any{
		"issue_id":   kit.IssueID,
		"posts":      kit.Posts,
		"card":       kit.Card.Lines(),
		"card_image": "/v1/public/issue/" + v.IssueID + "/card",
		// 9:16, for WhatsApp Status and Instagram Stories.
		"card_story": "/v1/public/issue/" + v.IssueID + "/card?shape=story",
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
	if polishMeta != nil {
		body["polish"] = polishMeta
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, body)
}

// polishKit rewords every post in the kit and keeps the deterministic one
// wherever the rewording cannot be proved to say the same facts.
//
// Three properties matter here and none of them is about the model.
//
// It runs the posts concurrently under one deadline. This is a public GET on
// a page load, and six calls in sequence at a per-call timeout would make the
// share kit slower than the thing it is trying to make easier.
//
// A refusal is logged and not surfaced as an error, because the response is
// already correct without it. The log line is the measurement: it names the
// post, the model and the prompt version, so a prompt that starts inventing
// numbers shows up in the logs rather than in a forwarded WhatsApp message.
//
// It never touches Review. A model can make a Marathi sentence fluent; the
// claim that a Marathi speaker has read it is a different claim and stays
// false until one has.
func (s *Server) polishKit(ctx context.Context, f share.Facts, kit *share.Kit, issueID string) map[string]any {
	// The whole kit, not each post: a slow provider must not multiply by the
	// number of languages.
	ctx, cancel := context.WithTimeout(ctx, shareKitPolishBudget)
	defer cancel()

	type outcome struct {
		key string
		res polish.Result
	}
	results := make(chan outcome, len(kit.Posts))
	var wg sync.WaitGroup
	for key, p := range kit.Posts {
		wg.Add(1)
		go func(key string, p share.Post) {
			defer wg.Done()
			results <- outcome{key, polish.Apply(ctx, s.polish, f, p)}
		}(key, p)
	}
	wg.Wait()
	close(results)

	var model, prompt string
	var worded, kept int
	for o := range results {
		if o.res.Model != "" {
			model, prompt = o.res.Model, o.res.Prompt
		}
		kit.Posts[o.key] = o.res.Post
		if len(o.res.Faults) == 0 {
			if o.res.Post.Wording == share.WordedByModel {
				worded++
			}
			continue
		}
		kept++
		s.log.Warn("kept the composed post rather than the model's wording",
			"issue_id", issueID, "post", o.key, "model", o.res.Model,
			"prompt", o.res.Prompt, "faults", len(o.res.Faults),
			"first_fault", o.res.Faults[0].String())
	}
	return map[string]any{
		"model": model, "prompt": prompt,
		"worded_by_model": worded, "kept_composed": kept,
		"note": "A model reworded the sentences the platform composed. It was " +
			"given no facts and no link, and every number, reference and " +
			"account in a reworded line was checked against the line it " +
			"replaced. Where the check failed, the composed text is what you " +
			"see.",
	}
}

// shareKitPolishBudget is the whole kit's share of a page load. A share kit
// that arrives late is a share that does not happen, and the composed post is
// already correct.
const shareKitPolishBudget = 10 * time.Second

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
	f, _ := s.factsOnly(r, v)

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

// factsOnly narrows the public view without consulting the registry.
//
// Separate from factsFor because the unfurl needs the same narrowing and no
// handles: metadata tags nobody, and a page render should not wait on a
// registry query it has no use for.
func (s *Server) factsOnly(r *http.Request, v store.PublicIssueView) (share.Facts, publish.Derivative) {
	d := publish.Derive(publish.Capture{
		ReportID: v.IssueID, Lat: v.Lat, Lon: v.Lon,
		Category: v.Category, Subcategory: v.Subcategory,
	}, v.Redaction)

	f := share.Facts{
		IssueID: v.IssueID,
		// A relative link, made absolute against the host serving this
		// request: the permalink must be the page the reader can open, and
		// hard-coding a domain here is how a staging link ends up forwarded.
		Permalink:        s.permalink(r, v.IssueID),
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
	return f, d
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

	// An unknown shape is refused rather than quietly served as the default:
	// somebody asking for a frame is going to post the result into it.
	shape := share.Feed
	if q := r.URL.Query().Get("shape"); q != "" {
		shape = share.Shape(q)
		if !share.KnownShape(shape) {
			writeError(w, http.StatusBadRequest,
				`shape must be "feed" or "story"`)
			return
		}
	}

	f, _ := s.factsFor(r, v)
	body, err := share.EncodeShape(photo, share.CardFor(f), shape)
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

// permalink is the canonical address of an issue page.
//
// It prefers the configured base URL and falls back to the request's own host
// only when none is configured. That order is the whole point, and it is a
// correctness rule rather than a tidiness one.
//
// `Host` is a request header. The share kit is a public, unauthenticated GET,
// so a caller chooses it: a request carrying `Host: evil.example` used to
// make the platform generate a post reading "Every fact above is on this
// page, with the document it came from:" above a link to a host the caller
// controls — in all three languages, and burned into the card image as well —
// ready to be forwarded past any page this platform can correct. Nothing
// downstream can tell that post from a real one.
//
// The second reason is slower and just as permanent. A forwarded post is
// uncorrectable, so a permalink that follows whatever hostname answered today
// dies the day the service's hostname changes. A configured canonical base
// survives that; a header does not.
//
// The fallback stays because a local run has no canonical URL and refusing
// would break development. cmd/api logs what it costs when PUBLIC_BASE_URL is
// unset.
func (s *Server) permalink(r *http.Request, issueID string) string {
	if s.publicBaseURL != "" {
		return strings.TrimSuffix(s.publicBaseURL, "/") + "/issue/" + issueID
	}
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
