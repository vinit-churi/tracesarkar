package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/redact"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Truth is the persistence the annotation surface needs.
//
// It exists because the redaction detector's recall was unmeasurable: every
// capture it had been run over happened to contain no face and no plate, so it
// had never had the opportunity to miss one. ADR 0008 required an adversarial
// set before launch — faces at distance, partial occlusion, reflections,
// plates at angle, night, Indian plate formats — and this is where a person
// marks one up.
type Truth interface {
	MediaAwaitingTruth(ctx context.Context, accountID string, limit int) ([]store.PendingTruth, error)
	SaveRedactionTruth(ctx context.Context, in store.RedactionTruth) error
	// MediaByID returns where a photograph is and who owns it, so a capture is
	// never served or annotated by anyone else.
	MediaByID(ctx context.Context, mediaID string) (store.MediaRef, error)
}

// truthItem is one photograph waiting to be marked up. The archive key is
// absent: it is internal, and a page that held it could read the private
// bucket directly.
type truthItem struct {
	MediaID    string `json:"media_id"`
	ReportID   string `json:"report_id"`
	CapturedAt string `json:"captured_at"`
	MediaURL   string `json:"media_url"`
}

// handleTruthQueue lists the account's own photographs that nobody has marked
// up, oldest first.
func (s *Server) handleTruthQueue(w http.ResponseWriter, r *http.Request) {
	if s.truth == nil {
		writeError(w, http.StatusNotImplemented, "annotation is not configured on this server")
		return
	}

	pending, err := s.truth.MediaAwaitingTruth(r.Context(), s.reporter(r), 200)
	if err != nil {
		s.log.Error("could not read the annotation queue", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read the annotation queue")
		return
	}

	items := make([]truthItem, 0, len(pending))
	for _, p := range pending {
		items = append(items, truthItem{
			MediaID:    p.MediaID,
			ReportID:   p.ReportID,
			CapturedAt: p.CapturedAt.UTC().Format(time.RFC3339),
			MediaURL:   "/v1/media/" + p.MediaID,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		// Served rather than copied into the page. A condition typed into the
		// page and not known here would split its own row in the breakdown,
		// which is the same mistake the field kit made with the taxonomy.
		"conditions": redact.Conditions(),
		"kinds":      []redact.Kind{redact.Face, redact.Plate},
	})
}

type truthRequest struct {
	Boxes         []redact.Box `json:"boxes"`
	PeoplePresent bool         `json:"people_present"`
	Conditions    []string     `json:"conditions"`
	Notes         string       `json:"notes"`
}

// handleSaveTruth records what a person can see in one photograph.
//
// An empty box list is a judgement, not an empty request: it says somebody
// looked and there is nothing to find, which is the only kind of case that can
// measure a false positive.
func (s *Server) handleSaveTruth(w http.ResponseWriter, r *http.Request) {
	if s.truth == nil {
		writeError(w, http.StatusNotImplemented, "annotation is not configured on this server")
		return
	}

	mediaID := r.PathValue("id")
	ref, err := s.truth.MediaByID(r.Context(), mediaID)
	// Another account's capture is "not found", not "forbidden": a 403 would
	// confirm that the photograph exists.
	if errors.Is(err, store.ErrNotFound) || (err == nil && ref.Owner != s.reporter(r)) {
		writeError(w, http.StatusNotFound, "no such photograph")
		return
	}
	if err != nil {
		s.log.Error("could not look up media", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that capture")
		return
	}

	var in truthRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not readable JSON")
		return
	}

	// Refused here, with the reason, rather than stored and averaged into a
	// figure later. A box reaching past the edge of the photograph or a
	// condition outside the vocabulary distorts quietly.
	for _, b := range in.Boxes {
		if err := b.Valid(); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if b.Kind == redact.Face && !in.PeoplePresent {
			writeError(w, http.StatusBadRequest,
				"a face is marked up but the judgement says nobody is present")
			return
		}
	}
	for _, c := range in.Conditions {
		if !redact.KnownCondition(c) {
			writeError(w, http.StatusBadRequest,
				"\""+c+"\" is not a condition the breakdown knows")
			return
		}
	}

	err = s.truth.SaveRedactionTruth(r.Context(), store.RedactionTruth{
		MediaID: mediaID,
		// Taken from the photograph's own row rather than the request body: a
		// report id a caller could choose would let a judgement be filed
		// against somebody else's capture.
		ReportID:      ref.ReportID,
		Boxes:         in.Boxes,
		PeoplePresent: in.PeoplePresent,
		Conditions:    in.Conditions,
		Notes:         in.Notes,
		AnnotatedBy:   s.reporter(r),
	})
	if err != nil {
		s.log.Error("could not record ground truth", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not record that judgement")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"media_id": mediaID,
		"faces":    count(in.Boxes, redact.Face),
		"plates":   count(in.Boxes, redact.Plate),
	})
}

func count(boxes []redact.Box, kind redact.Kind) int {
	n := 0
	for _, b := range boxes {
		if b.Kind == kind {
			n++
		}
	}
	return n
}

// handleMedia serves one photograph by its own id, to the account that took it.
//
// The archival original, which is where a face still is. The bucket is private,
// this is the only way to see one, and the annotator needs exactly this image:
// marking up the redacted copy would be marking up the answer.
func (s *Server) handleMedia(w http.ResponseWriter, r *http.Request) {
	if s.truth == nil || s.blobs == nil {
		writeError(w, http.StatusNotImplemented, "media is not configured on this server")
		return
	}

	ref, err := s.truth.MediaByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) || (err == nil && ref.Owner != s.reporter(r)) {
		writeError(w, http.StatusNotFound, "no such photograph")
		return
	}
	if err != nil {
		s.log.Error("could not look up media", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that capture")
		return
	}

	body, err := s.blobs.Get(r.Context(), ref.Key)
	if err != nil {
		s.log.Error("could not read media", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that capture")
		return
	}

	contentType := ref.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	// Private: a citizen's photograph, un-redacted, served to one person. A
	// shared cache must not keep a copy that outlives the session.
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
