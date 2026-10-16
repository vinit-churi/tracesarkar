package api

import (
	"context"
	"net/http"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/issue"
	"github.com/vinit-churi/tracesarkar/internal/publish"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// PublicIssues is the second surface served without a sign-in.
//
// The ward page was safe because nothing on it came from a citizen. This one
// is not safe for that reason — it is an account of somebody's report — so it
// is safe only by construction: the position is a cell, the history carries no
// actor, and the photograph exists only behind a completed redaction.
type PublicIssues interface {
	PublicIssue(ctx context.Context, id string) (store.PublicIssueView, bool, error)
}

// statusLabels says what each status means in words a resident would use.
//
// claimed_resolved and citizen_confirmed must never share a label or a count.
// They are different claims by different parties, and collapsing them is how a
// platform ends up reporting an authority's own say-so as a fix.
var statusLabels = map[string]string{
	"open":              "Open",
	"claimed_resolved":  "The authority has recorded this as attended",
	"citizen_confirmed": "A resident photographed it fixed",
	"reopened":          "Seen again after it was recorded as attended",
}

// handlePublicIssue serves one issue to anybody.
func (s *Server) handlePublicIssue(w http.ResponseWriter, r *http.Request) {
	v, ok := s.lookUpPublicIssue(w, r)
	if !ok {
		return
	}

	// publish.Derive is the gate. It coarsens the position and refuses to name
	// an image without a completed redaction behind it, so this handler never
	// decides either question for itself.
	d := publish.Derive(publish.Capture{
		ReportID:    v.IssueID,
		Lat:         v.Lat,
		Lon:         v.Lon,
		Category:    v.Category,
		Subcategory: v.Subcategory,
	}, v.Redaction)

	body := map[string]any{
		"issue_id":    v.IssueID,
		"authority":   v.Authority,
		"ward":        v.Ward,
		"category":    d.Category,
		"subcategory": d.Subcategory,
		"status":      v.Status,
		// Written out rather than left as a code, because a public page that
		// prints `claimed_resolved` has told the reader nothing.
		"status_label": label(v.Status),
		"at": map[string]any{
			"lat":         d.Lat,
			"lon":         d.Lon,
			"cell_metres": d.CellMetres,
			"note": "This is a cell, not a point. The exact position of a " +
				"report is never published.",
		},
		"first_seen": v.FirstSeen.UTC().Format("2 January 2006"),
		"last_seen":  v.LastSeen.UTC().Format("2 January 2006"),
		"history":    publicHistory(v.Events),
	}

	// D121: the line appears from the second sighting onward. "1 person has
	// reported this" claims company the issue does not have.
	if v.Sightings > 1 {
		body["corroboration"] = map[string]any{
			"sightings": v.Sightings,
			"note":      "One problem, seen several times. Not several complaints.",
		}
	}

	if d.ImageKey != "" {
		// A path on this server, never the object key and never a tokenised
		// URL: the gate has to sit in front of every read, not beside one.
		body["image"] = "/v1/public/issue/" + v.IssueID + "/image"
		body["image_note"] = "Faces and number plates are obscured before this " +
			"image exists."
	} else {
		body["image_withheld"] = d.Withheld
	}

	addLodging(body, v)
	addProcess(body, v)

	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, body)
}

// addLodging states what is known about the authority having been told.
//
// The clock runs from the lodging and from nothing else, and it is shown only
// with the document that creates it: a deadline with no citation is an
// assertion, and this platform does not make those.
func addLodging(body map[string]any, v store.PublicIssueView) {
	if v.NotifiedAt == nil {
		body["lodged"] = map[string]any{
			"state": "not_lodged",
			"note": "Nobody has recorded telling the authority about this yet, " +
				"so no deadline is running. Photographing a problem does not " +
				"notify anybody.",
		}
		return
	}

	lodged := map[string]any{
		"state": "lodged",
		"on":    v.NotifiedAt.UTC().Format("2 January 2006"),
	}
	if v.DeadlineHours > 0 && v.DeadlineCitation != "" {
		reading := issue.NewClock(v.NotifiedAt, v.DeadlineHours).At(time.Now().UTC())
		lodged["deadline_hours"] = v.DeadlineHours
		lodged["deadline_citation"] = v.DeadlineCitation
		lodged["deadline_state"] = string(reading.State)
		lodged["deadline_says"] = reading.Says
	}
	body["lodged"] = lodged
}

// addProcess states how the thing photographed is lawfully done.
//
// It sits apart from the clock on purpose. The Tree Authority's forty-five
// days are the time it has to decide an application to fell a tree, not the
// time it has to answer a report that one came down (D160), and a figure in
// the wrong box would say the second. Each note carries the document that
// publishes it, and a note without one never reaches here.
func addProcess(body map[string]any, v store.PublicIssueView) {
	if len(v.ProcessNotes) == 0 {
		return
	}
	notes := make([]map[string]any, 0, len(v.ProcessNotes))
	for _, n := range v.ProcessNotes {
		notes = append(notes, map[string]any{
			"says": n.Says,
			"cite": n.Citation,
		})
	}
	body["process"] = map[string]any{
		"heading": "What the law requires for this",
		"note": "These are published figures about how this is lawfully done. " +
			"They are not a deadline for answering a report, and they say " +
			"nothing about whether anybody followed them.",
		"facts": notes,
	}
}

func label(status string) string {
	if l, ok := statusLabels[status]; ok {
		return l
	}
	// An unknown status is reported as unknown rather than guessed at, so a
	// new status added to the database cannot quietly inherit another's
	// meaning on a public page.
	return "Status not published"
}

func publicHistory(events []store.PublicEvent) []map[string]any {
	out := make([]map[string]any, 0, len(events))
	for _, e := range events {
		out = append(out, map[string]any{
			"seq":    e.Seq,
			"kind":   e.Kind,
			"on":     e.At.UTC().Format("2 January 2006"),
			"detail": e.Detail,
			"hash":   e.Hash,
		})
	}
	return out
}

// handlePublicIssueImage streams the redacted copy.
//
// Every read goes through publish.Derive first, so a photograph the scan held
// is a 404 here even though the object exists and this server could read it.
// The object store is not touched until the gate has said yes.
func (s *Server) handlePublicIssueImage(w http.ResponseWriter, r *http.Request) {
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

	// The stored copy is full size, because that is what the classifier reads.
	// A public page is costed at 150 KB on a 2G connection, and a page most of
	// the people this is for cannot load is not a public page.
	body, err := publish.ForWeb(stored)
	if err != nil {
		s.log.Error("could not size a photograph for the web", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that photograph")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	// Public, because every face and plate in it has been obscured. The
	// unredacted original is served by a different route, to its owner only.
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (s *Server) lookUpPublicIssue(w http.ResponseWriter, r *http.Request) (store.PublicIssueView, bool) {
	if s.publicIssues == nil {
		writeError(w, http.StatusNotImplemented, "public issues are not configured on this server")
		return store.PublicIssueView{}, false
	}
	v, found, err := s.publicIssues.PublicIssue(r.Context(), r.PathValue("id"))
	if err != nil {
		s.log.Error("could not read a public issue", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that issue")
		return store.PublicIssueView{}, false
	}
	if !found {
		writeError(w, http.StatusNotFound, "no such issue")
		return store.PublicIssueView{}, false
	}
	return v, true
}
