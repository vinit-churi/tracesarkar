package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/vinit-churi/tracesarkar/internal/attribute"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// ReviewItem is the store's queued answer, shared so handlers need no
// translation layer.
type ReviewItem = store.ReviewItem

// Reviews is the persistence the attribution review surface needs.
type Reviews interface {
	PendingReviewItems(ctx context.Context, limit int) ([]ReviewItem, error)
	RecordVerdict(ctx context.Context, id, verdict, note, reviewer string) error
	ReviewSummary(ctx context.Context) (attribute.Summary, error)
	// ExpectationAgreement reports whether the join returned the segment a
	// probe was generated from. Read only after a verdict is cast.
	ExpectationAgreement(ctx context.Context, id string) (*bool, error)
}

// handleReviewQueue lists answers waiting for a verdict.
func (s *Server) handleReviewQueue(w http.ResponseWriter, r *http.Request) {
	if s.reviews == nil {
		writeError(w, http.StatusNotImplemented, "the review queue is not configured on this server")
		return
	}

	items, err := s.reviews.PendingReviewItems(r.Context(), 50)
	if err != nil {
		s.log.Error("could not read review queue", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read the review queue")
		return
	}
	summary, err := s.reviews.ReviewSummary(r.Context())
	if err != nil {
		s.log.Error("could not summarise reviews", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read the review queue")
		return
	}

	if items == nil {
		items = []ReviewItem{}
	}
	// The reviewer must not be told the expected answer before they give
	// theirs. Showing it leads the witness: they agree with the hint rather
	// than with the map, and the precision number stops being independent
	// evidence of anything. It is reported back after the verdict instead.
	for i := range items {
		items[i].AgreesWithExpectation = nil
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":   items,
		"summary": summary,
		// Stated rather than computed in the client, so the number shown and
		// the number recorded cannot drift apart.
		"has_precision": summary.HasPrecision(),
		"precision":     summary.Precision(),
	})
}

// handleReviewVerdict records what a person decided about one answer.
func (s *Server) handleReviewVerdict(w http.ResponseWriter, r *http.Request) {
	if s.reviews == nil {
		writeError(w, http.StatusNotImplemented, "the review queue is not configured on this server")
		return
	}

	// A verdict is a judgement, and a precision number is only as good as
	// knowing whose judgement produced it. The static field-kit token names
	// nobody, so it cannot cast one.
	claims, ok := claimsFrom(r.Context())
	if !ok || claims.AccountID == "" {
		writeError(w, http.StatusForbidden,
			"a verdict must be recorded against a signed-in reviewer")
		return
	}

	var in struct {
		Verdict string `json:"verdict"`
		Note    string `json:"note"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not valid JSON")
		return
	}
	switch in.Verdict {
	case "correct", "wrong", "unsure":
	default:
		writeError(w, http.StatusBadRequest,
			"verdict must be one of: correct, wrong, unsure")
		return
	}

	id := r.PathValue("id")
	if err := s.reviews.RecordVerdict(r.Context(), id, in.Verdict, in.Note, claims.AccountID); err != nil {
		s.log.Error("could not record verdict", "error", err.Error(), "item", id)
		writeError(w, http.StatusInternalServerError, "could not record the verdict")
		return
	}

	out := map[string]any{"recorded": true}

	// Now it is safe, and useful: a disagreement between the person and the
	// probe is exactly the case worth looking into.
	if agreed, err := s.reviews.ExpectationAgreement(r.Context(), id); err == nil && agreed != nil {
		out["agreed_with_expectation"] = *agreed
	}

	summary, err := s.reviews.ReviewSummary(r.Context())
	if err == nil {
		out["summary"] = summary
		out["has_precision"] = summary.HasPrecision()
		out["precision"] = summary.Precision()
	}
	writeJSON(w, http.StatusOK, out)
}
