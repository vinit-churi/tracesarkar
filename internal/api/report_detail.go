package api

import (
	"context"
	"net/http"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

// ReportDetail is the store's view of one capture, shared so handlers need no
// translation layer.
type ReportDetail = store.ReportDetail

// Details reads one capture and everything concluded about it.
type Details interface {
	ReportDetail(ctx context.Context, id string) (ReportDetail, bool, error)
}

// handleReportDetail returns one capture to the person who made it.
func (s *Server) handleReportDetail(w http.ResponseWriter, r *http.Request) {
	if s.details == nil {
		writeError(w, http.StatusNotImplemented, "report detail is not configured on this server")
		return
	}

	// The shared field-kit token names nobody, so it cannot be said to own a
	// report. Only a signed-in person can read one.
	claims, ok := claimsFrom(r.Context())
	if !ok || claims.AccountID == "" {
		writeError(w, http.StatusForbidden, "sign in to read a report")
		return
	}

	detail, found, err := s.details.ReportDetail(r.Context(), r.PathValue("id"))
	if err != nil {
		s.log.Error("could not read report", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read the report")
		return
	}

	// A report carries a precise position and a photograph of somewhere a
	// person stood. Someone else's is not public because the id is guessable,
	// and answering 404 rather than 403 avoids confirming it exists at all.
	if !found || detail.AccountID != claims.AccountID {
		writeError(w, http.StatusNotFound, "no such report")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"report": detail})
}
