package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/classify"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Labels is the persistence the labelling surface needs.
//
// Labelling is what turns a day of walking into an evaluation set. Until it
// exists, system 3 has a classifier and no way to say whether it is any good.
type Labels interface {
	UnlabelledReports(ctx context.Context, limit int) ([]store.PendingLabel, error)
	SaveReportLabel(ctx context.Context, in store.ReportLabel) error
	// MediaFor returns the photograph's key and the account that owns it, so a
	// capture is never served to anyone else.
	MediaFor(ctx context.Context, reportID string) (key, contentType, owner string, err error)
}

// Blobs reads stored media back. Separate from Media, which only writes: the
// capture path and the labelling path want opposite halves of the same store.
type Blobs interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

// queueItem is deliberately thin. It carries what is needed to show a
// photograph and nothing that could tell the labeller what the classifier
// thought — the archive key included, which is internal and would let a page
// read the bucket directly.
type queueItem struct {
	ReportID   string `json:"report_id"`
	Ward       string `json:"ward,omitempty"`
	CapturedAt string `json:"captured_at"`
	MediaURL   string `json:"media_url"`
}

// handleLabelQueue lists captures waiting for a human judgement.
func (s *Server) handleLabelQueue(w http.ResponseWriter, r *http.Request) {
	if s.labels == nil {
		writeError(w, http.StatusNotImplemented, "labelling is not configured on this server")
		return
	}

	pending, err := s.labels.UnlabelledReports(r.Context(), 200)
	if err != nil {
		s.log.Error("could not read the labelling queue", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read the labelling queue")
		return
	}

	items := make([]queueItem, 0, len(pending))
	for _, p := range pending {
		items = append(items, queueItem{
			ReportID:   p.ReportID,
			Ward:       p.Ward,
			CapturedAt: p.CapturedAt.UTC().Format(time.RFC3339),
			MediaURL:   "/v1/reports/" + p.ReportID + "/media",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":    items,
		"taxonomy": classify.Taxonomy(),
	})
}

type labelRequest struct {
	Label      string   `json:"label"`
	FrameType  string   `json:"frame_type"`
	Conditions []string `json:"conditions"`
	Ward       string   `json:"ward_ground_truth"`
	Notes      string   `json:"notes"`
}

// handleSaveLabel records one human judgement.
func (s *Server) handleSaveLabel(w http.ResponseWriter, r *http.Request) {
	if s.labels == nil {
		writeError(w, http.StatusNotImplemented, "labelling is not configured on this server")
		return
	}

	var in labelRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not readable JSON")
		return
	}

	// Only a subcategory the taxonomy knows. Free text would make the eval set
	// unscoreable: nothing downstream could decide whether the classifier's
	// answer matched it, and a typo would read as a miss forever.
	label := strings.ToLower(strings.TrimSpace(in.Label))
	if classify.CategoryOf(label) == "" {
		writeError(w, http.StatusBadRequest,
			"the label must be one of the taxonomy's subcategories")
		return
	}

	frame := strings.TrimSpace(in.FrameType)
	if frame == "" {
		frame = "close"
	}

	if err := s.labels.SaveReportLabel(r.Context(), store.ReportLabel{
		ReportID:        r.PathValue("id"),
		FrameType:       frame,
		Label:           label,
		Conditions:      in.Conditions,
		WardGroundTruth: strings.TrimSpace(in.Ward),
		Notes:           strings.TrimSpace(in.Notes),
		LabelledBy:      s.reporter(r),
	}); err != nil {
		s.log.Error("could not save label", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not save that label")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"label":    label,
		"category": classify.CategoryOf(label),
		"hazard":   classify.IsHazard(label),
	})
}

// handleReportMedia serves one capture's photograph.
//
// The archive bucket is not public — it holds photographs taken by people who
// did not agree to publish them — so this is the only way to see one, and it
// serves a capture only to the account that made it.
func (s *Server) handleReportMedia(w http.ResponseWriter, r *http.Request) {
	if s.labels == nil || s.blobs == nil {
		writeError(w, http.StatusNotImplemented, "media is not configured on this server")
		return
	}

	key, contentType, owner, err := s.labels.MediaFor(r.Context(), r.PathValue("id"))
	// Another account's capture is "not found", not "forbidden": a 403 would
	// confirm the report exists, which is itself a disclosure.
	if errors.Is(err, store.ErrNotFound) || (err == nil && owner != s.reporter(r)) {
		writeError(w, http.StatusNotFound, "no such report")
		return
	}
	if err != nil {
		s.log.Error("could not look up media", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that capture")
		return
	}

	body, err := s.blobs.Get(r.Context(), key)
	if err != nil {
		s.log.Error("could not read media", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that capture")
		return
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	// Private: this is a citizen's photograph, and a shared cache must not
	// keep a copy that outlives the session that asked for it.
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
