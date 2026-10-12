package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/issue"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Issues is the persistence the issue surface needs.
type Issues interface {
	IssueForReport(ctx context.Context, reportID string) (store.IssueView, bool, error)
	RecordNotified(ctx context.Context, issueID, reference, actor string, at time.Time) error
	ReporterOf(ctx context.Context, issueID, accountID string) (bool, error)
}

type lodgedRequest struct {
	Reference string `json:"reference"`
}

// handleLodged records that the authority has been told about an issue.
//
// This is the only thing that starts a clock. Everything else the platform
// does is description; this is the moment an obligation begins, which is why
// it is a deliberate act by the person who reported it rather than something
// inferred.
func (s *Server) handleLodged(w http.ResponseWriter, r *http.Request) {
	if s.issues == nil {
		writeError(w, http.StatusNotImplemented, "issues are not configured on this server")
		return
	}

	issueID := r.PathValue("id")
	// Only somebody who photographed it. The reference dates a legal clock,
	// and that is not something to assert about another person's report.
	ok, err := s.issues.ReporterOf(r.Context(), issueID, s.reporter(r))
	if err != nil {
		s.log.Error("could not check reporter", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not record that")
		return
	}
	if !ok {
		// Not found rather than forbidden: a 403 confirms the issue exists.
		writeError(w, http.StatusNotFound, "no such issue")
		return
	}

	var in lodgedRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not readable JSON")
		return
	}

	err = s.issues.RecordNotified(r.Context(), issueID,
		strings.TrimSpace(in.Reference), s.reporter(r), time.Now().UTC())
	switch {
	case errors.Is(err, store.ErrAlreadyNotified):
		// Refused rather than overwritten: moving the start would move the
		// deadline, and a clock that can be restarted means nothing.
		writeError(w, http.StatusConflict,
			"This was already recorded as lodged. The deadline runs from the "+
				"first time it was reported to the authority, so it is not moved.")
		return
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "no such issue")
		return
	case err != nil:
		s.log.Error("could not record lodging", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not record that")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lodged": true})
}

// issuePayload is what a citizen sees about the problem behind their capture.
type issuePayload struct {
	IssueID   string `json:"issue_id"`
	Status    string `json:"status"`
	Sightings int    `json:"sightings"`
	Reference string `json:"reference,omitempty"`
	Lodged    bool   `json:"lodged"`
	// Clock is rendered from legal_constants, never from a literal.
	ClockState string `json:"clock_state"`
	ClockSays  string `json:"clock_says"`
}

// issueFor builds that payload, or nil when the capture is not yet a sighting
// of anything — classification has to run before a category exists, and the
// category is what decides which issue it belongs to.
func (s *Server) issueFor(ctx context.Context, reportID string, deadlineHours int) *issuePayload {
	if s.issues == nil {
		return nil
	}
	v, found, err := s.issues.IssueForReport(ctx, reportID)
	if err != nil || !found {
		return nil
	}
	reading := issue.NewClock(v.NotifiedAt, deadlineHours).At(time.Now().UTC())
	return &issuePayload{
		IssueID:    v.IssueID,
		Status:     v.Status,
		Sightings:  v.Sightings,
		Reference:  v.Reference,
		Lodged:     v.NotifiedAt != nil,
		ClockState: string(reading.State),
		ClockSays:  reading.Says,
	}
}
