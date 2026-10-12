package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

type fakeIssues struct {
	view     store.IssueView
	found    bool
	reporter string
	notified error
	recorded int
}

func (f *fakeIssues) IssueForReport(context.Context, string) (store.IssueView, bool, error) {
	return f.view, f.found, nil
}
func (f *fakeIssues) RecordNotified(_ context.Context, _, _, _ string, _ time.Time) error {
	if f.notified != nil {
		return f.notified
	}
	f.recorded++
	return nil
}
func (f *fakeIssues) ReporterOf(_ context.Context, _, accountID string) (bool, error) {
	return accountID == f.reporter, nil
}

func issueServer(t *testing.T, issues *fakeIssues) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{}, Issues: issues,
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func lodge(t *testing.T, srv *Server, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", "/v1/issues/i1/lodged", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestLodgingStartsTheClock(t *testing.T) {
	issues := &fakeIssues{reporter: "field-kit"}
	if rec := lodge(t, issueServer(t, issues), `{"reference":"MARG-1"}`); rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	if issues.recorded != 1 {
		t.Error("the lodging was not recorded")
	}
}

// Moving the start would move the deadline. A clock that can be restarted by
// re-submitting is a clock that means nothing.
func TestASecondLodgingDoesNotMoveTheDeadline(t *testing.T) {
	issues := &fakeIssues{reporter: "field-kit", notified: store.ErrAlreadyNotified}
	rec := lodge(t, issueServer(t, issues), `{"reference":"MARG-2"}`)

	if rec.Code != http.StatusConflict {
		t.Fatalf("got %d, want 409", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "not moved") {
		t.Errorf("the refusal does not explain why: %s", rec.Body.String())
	}
}

// The reference dates a legal clock. Somebody who never photographed it does
// not get to assert that about another person's report — and is told the issue
// does not exist rather than that they may not touch it.
func TestOnlyAReporterMayLodge(t *testing.T) {
	issues := &fakeIssues{reporter: "somebody-else"}
	rec := lodge(t, issueServer(t, issues), `{"reference":"MARG-3"}`)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rec.Code)
	}
	if issues.recorded != 0 {
		t.Error("a stranger's lodging was recorded")
	}
}
