package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/vinit-churi/tracesarkar/internal/classify"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Errors the confirmation handlers distinguish.
var (
	ErrNotAwaitingConfirmation = store.ErrNotAwaitingConfirmation
	ErrNotAnOfferedOption      = store.ErrNotAnOfferedOption
)

// Confirmations is the persistence the confirmation endpoints need.
type Confirmations interface {
	CapturesAwaitingConfirmation(ctx context.Context, accountID string, limit int) ([]store.AwaitingConfirmation, error)
	ConfirmClassification(ctx context.Context, reportID, accountID, chosen string) (classify.Decision, error)
}

// handleAwaitingConfirmation lists the questions this account has not
// answered.
//
// The platform asks when the classifier named a subcategory it cannot route,
// or a category with no department behind it, or scored below the confidence
// floor. Until this endpoint existed the question was computed and discarded
// and the capture waited for ever, which is hard rule 7's failure mode with
// the photograph still safely stored.
func (s *Server) handleAwaitingConfirmation(w http.ResponseWriter, r *http.Request) {
	if s.confirmations == nil {
		writeError(w, http.StatusNotImplemented, "confirmations are not configured on this server")
		return
	}
	pending, err := s.confirmations.CapturesAwaitingConfirmation(r.Context(), s.reporter(r), 50)
	if err != nil {
		s.log.Error("could not read captures awaiting confirmation", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read your captures")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"captures": pending,
		"note": "The platform could not tell what these are from the photograph " +
			"alone. Nothing is sent anywhere until you answer; answering is what " +
			"lets the report be routed.",
	})
}

// handleConfirmClassification records the answer.
func (s *Server) handleConfirmClassification(w http.ResponseWriter, r *http.Request) {
	if s.confirmations == nil {
		writeError(w, http.StatusNotImplemented, "confirmations are not configured on this server")
		return
	}
	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		writeError(w, http.StatusBadRequest, "a capture id is required")
		return
	}

	var in struct {
		Chosen string `json:"chosen"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not valid JSON")
		return
	}
	if strings.TrimSpace(in.Chosen) == "" {
		writeError(w, http.StatusBadRequest, "chosen is required")
		return
	}

	dec, err := s.confirmations.ConfirmClassification(r.Context(), id, s.reporter(r), in.Chosen)
	switch {
	case errors.Is(err, ErrNotAwaitingConfirmation):
		// 404 for all four of "no such capture", "somebody else's capture",
		// "never asked about" and "already answered". A 403 on somebody
		// else's id would confirm that the id exists, which is how a list of
		// other people's captures gets enumerated.
		writeError(w, http.StatusNotFound, "that capture is not waiting for an answer")
		return
	case errors.Is(err, ErrNotAnOfferedOption):
		writeError(w, http.StatusBadRequest,
			"that is not one of the options offered for this capture")
		return
	case err != nil:
		s.log.Error("could not record a confirmation", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not record your answer")
		return
	}

	s.log.Info("reporter confirmed a classification", "report_id", id,
		"category", dec.Category, "subcategory", dec.Subcategory,
		"outcome", string(dec.Outcome))

	body := map[string]any{
		"status":      "recorded",
		"category":    dec.Category,
		"subcategory": dec.Subcategory,
		"outcome":     string(dec.Outcome),
	}
	// Two honest endings, and they must not read the same. An answer the
	// platform can route becomes a report; an answer in a category with no
	// department behind it is recorded and goes no further, and saying
	// otherwise would be the promise this platform exists not to make.
	if dec.Outcome == classify.NotYetCovered {
		body["note"] = "Recorded. Nobody is routed for this kind of problem in this " +
			"ward yet, so this will not become a report until a department is on " +
			"record for it. Your photograph and your answer are kept."
	} else {
		body["note"] = "Recorded. This will be matched to a department and joined " +
			"to any earlier sightings of the same problem on the next pass."
	}
	writeJSON(w, http.StatusOK, body)
}
