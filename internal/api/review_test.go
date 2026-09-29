package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/attribute"
	"github.com/vinit-churi/tracesarkar/internal/auth"
)

type fakeReviews struct {
	pending  []ReviewItem
	summary  attribute.Summary
	verdicts []recordedVerdict
	err      error
}

type recordedVerdict struct{ id, verdict, note, reviewer string }

func (f *fakeReviews) PendingReviewItems(_ context.Context, limit int) ([]ReviewItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	if limit > 0 && limit < len(f.pending) {
		return f.pending[:limit], nil
	}
	return f.pending, nil
}

func (f *fakeReviews) RecordVerdict(_ context.Context, id, verdict, note, reviewer string) error {
	if f.err != nil {
		return f.err
	}
	switch verdict {
	case "correct", "wrong", "unsure":
	default:
		return errors.New("bad verdict")
	}
	f.verdicts = append(f.verdicts, recordedVerdict{id, verdict, note, reviewer})
	return nil
}

func (f *fakeReviews) ReviewSummary(context.Context) (attribute.Summary, error) {
	return f.summary, f.err
}

func reviewServer(t *testing.T, reviews *fakeReviews) (*Server, string) {
	t.Helper()
	issuer, err := auth.NewIssuer("a-test-signing-secret-32-bytes!!!", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{}, Token: "test-token",
		Account: "account-1", Issuer: issuer, Reviews: reviews,
	})
	if err != nil {
		t.Fatal(err)
	}
	token, err := issuer.Issue("reviewer-1", "r@example.org")
	if err != nil {
		t.Fatal(err)
	}
	return srv, token
}

func TestReviewQueueNeedsASignedInReviewer(t *testing.T) {
	srv, _ := reviewServer(t, &fakeReviews{})

	req := httptest.NewRequest(http.MethodGet, "/v1/review/attribution", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestReviewQueueReturnsPendingItemsAndTheSummary(t *testing.T) {
	distance := 4.2
	reviews := &fakeReviews{
		pending: []ReviewItem{{
			ID: "item-1", Kind: "probe", Lat: 19.23, Lon: 72.84, AccuracyM: 8,
			Confidence: "high", MatchedWorkCode: "W-447",
			MatchedContractor: "M/s Example Infracon Pvt. Ltd",
			MatchedLocation:   "Jain Mandir Road", DistanceM: &distance,
			Basis: "The point lies 4.2 m from Jain Mandir Road.",
		}},
		summary: attribute.Summary{Total: 50, Correct: 8, Wrong: 1, Pending: 41},
	}
	srv, token := reviewServer(t, reviews)

	req := httptest.NewRequest(http.MethodGet, "/v1/review/attribution", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Items   []ReviewItem      `json:"items"`
		Summary attribute.Summary `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].MatchedContractor == "" {
		t.Fatalf("items: %+v", body.Items)
	}
	if body.Summary.Total != 50 || body.Summary.Correct != 8 {
		t.Errorf("summary: %+v", body.Summary)
	}
}

func TestAVerdictIsRecordedAgainstTheReviewer(t *testing.T) {
	reviews := &fakeReviews{}
	srv, token := reviewServer(t, reviews)

	req := httptest.NewRequest(http.MethodPost, "/v1/review/attribution/item-1",
		strings.NewReader(`{"verdict":"wrong","note":"this is the service road"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	if len(reviews.verdicts) != 1 {
		t.Fatalf("verdicts: %+v", reviews.verdicts)
	}
	got := reviews.verdicts[0]
	if got.id != "item-1" || got.verdict != "wrong" {
		t.Errorf("recorded: %+v", got)
	}
	// Who judged it matters: a precision number is only as good as knowing
	// whose judgement produced it.
	if got.reviewer != "reviewer-1" {
		t.Errorf("reviewer: got %q, want %q", got.reviewer, "reviewer-1")
	}
	if got.note != "this is the service road" {
		t.Errorf("note lost: %q", got.note)
	}
}

func TestAnInventedVerdictIsRefused(t *testing.T) {
	reviews := &fakeReviews{}
	srv, token := reviewServer(t, reviews)

	req := httptest.NewRequest(http.MethodPost, "/v1/review/attribution/item-1",
		strings.NewReader(`{"verdict":"probably fine"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
	if len(reviews.verdicts) != 0 {
		t.Error("nothing should be recorded for an invalid verdict")
	}
}

func TestTheFieldKitTokenCannotCastAVerdict(t *testing.T) {
	// The static token names nobody. A precision number attributed to "the
	// field kit" is not a judgement anyone can stand behind.
	reviews := &fakeReviews{}
	srv, _ := reviewServer(t, reviews)

	req := httptest.NewRequest(http.MethodPost, "/v1/review/attribution/item-1",
		strings.NewReader(`{"verdict":"correct"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("got %d, want 403", rec.Code)
	}
	if len(reviews.verdicts) != 0 {
		t.Error("an anonymous caller must not record a verdict")
	}
}
