package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/classify"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

type fakeConfirmations struct {
	pending  []store.AwaitingConfirmation
	listErr  error
	decision classify.Decision
	err      error
	// seen records what the handler asked for, so a test can assert the
	// scoping rather than trust it.
	seenAccount, seenReport, seenChosen string
}

func (f *fakeConfirmations) CapturesAwaitingConfirmation(_ context.Context, accountID string, _ int) ([]store.AwaitingConfirmation, error) {
	f.seenAccount = accountID
	return f.pending, f.listErr
}

func (f *fakeConfirmations) ConfirmClassification(_ context.Context, reportID, accountID, chosen string) (classify.Decision, error) {
	f.seenReport, f.seenAccount, f.seenChosen = reportID, accountID, chosen
	return f.decision, f.err
}

func confServer(t *testing.T, c Confirmations) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Confirmations: c, Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func confDo(t *testing.T, srv *Server, method, path, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	r.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, r)
	var out map[string]any
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("body is not JSON: %v\n%s", err, rec.Body.String())
		}
	}
	return rec, out
}

// "awaiting-confirmation" sits under the same prefix as "{id}", and a
// wildcard winning here would send the list request to the detail handler.
func TestTheConfirmationListIsNotMistakenForACaptureID(t *testing.T) {
	c := &fakeConfirmations{pending: []store.AwaitingConfirmation{{
		ReportID: "r1", Category: "waste", Options: []string{"debris", "illegal dumping"},
	}}}
	rec, body := confDo(t, confServer(t, c), http.MethodGet,
		"/v1/reports/awaiting-confirmation", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if body["captures"] == nil {
		t.Errorf("the list handler did not answer: %s", rec.Body.String())
	}
	if c.seenAccount == "" {
		t.Error("the list was not scoped to an account")
	}
}

// Every refusal is a 404. A 403 on somebody else's id would confirm the id
// exists, which is how a list of other people's captures gets enumerated.
func TestConfirmingSomebodyElsesCaptureIsIndistinguishableFromAMissingOne(t *testing.T) {
	for _, err := range []error{ErrNotAwaitingConfirmation} {
		c := &fakeConfirmations{err: err}
		rec, _ := confDo(t, confServer(t, c), http.MethodPost,
			"/v1/reports/3f1a0f9c-0000-4000-8000-000000000001/confirm",
			`{"chosen":"pothole"}`)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%v answered %d, want 404", err, rec.Code)
		}
		if strings.Contains(strings.ToLower(rec.Body.String()), "forbidden") ||
			strings.Contains(rec.Body.String(), "another account") {
			t.Errorf("the refusal says whose capture it is: %s", rec.Body.String())
		}
	}
}

func TestConfirmingRefusesAnOptionThatWasNotOffered(t *testing.T) {
	c := &fakeConfirmations{err: ErrNotAnOfferedOption}
	rec, _ := confDo(t, confServer(t, c), http.MethodPost,
		"/v1/reports/r1/confirm", `{"chosen":"something else"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", rec.Code)
	}
}

func TestConfirmingValidatesTheBody(t *testing.T) {
	srv := confServer(t, &fakeConfirmations{})
	for _, body := range []string{``, `{}`, `{"chosen":""}`, `{"chosen":"   "}`, `not json`} {
		rec, _ := confDo(t, srv, http.MethodPost, "/v1/reports/r1/confirm", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %q answered %d, want 400", body, rec.Code)
		}
	}
}

// Two honest endings that must not read the same: an answer the platform can
// route becomes a report, and an answer in a category with no department
// behind it is recorded and goes no further. Saying otherwise would be the
// promise this platform exists not to make.
func TestAnAnswerNobodyRoutesDoesNotPromiseAReport(t *testing.T) {
	routable := &fakeConfirmations{decision: classify.Decision{
		Result:  classify.Result{Category: "waste", Subcategory: "debris"},
		Outcome: classify.Accepted}}
	_, body := confDo(t, confServer(t, routable), http.MethodPost,
		"/v1/reports/r1/confirm", `{"chosen":"debris"}`)
	routed, _ := body["note"].(string)

	uncovered := &fakeConfirmations{decision: classify.Decision{
		Result: classify.Result{Category: "environment",
			Subcategory: "mangrove destruction"},
		Outcome: classify.NotYetCovered}}
	_, body = confDo(t, confServer(t, uncovered), http.MethodPost,
		"/v1/reports/r1/confirm", `{"chosen":"mangrove destruction"}`)
	notRouted, _ := body["note"].(string)

	if routed == notRouted {
		t.Errorf("a routable answer and an unroutable one read the same:\n%s", routed)
	}
	if !strings.Contains(notRouted, "not become a report") {
		t.Errorf("the unroutable answer does not say so:\n%s", notRouted)
	}
	// And neither may claim anything was filed with an authority.
	for _, note := range []string{routed, notRouted} {
		for _, forbidden := range []string{"filed", "submitted", "sent to the"} {
			if strings.Contains(strings.ToLower(note), forbidden) {
				t.Errorf("the note claims the platform filed something: %q", note)
			}
		}
	}
}

func TestConfirmationNeedsASession(t *testing.T) {
	srv := confServer(t, &fakeConfirmations{})
	for _, path := range []string{"/v1/reports/awaiting-confirmation", "/v1/reports/r1/confirm"} {
		rec := httptest.NewRecorder()
		method := http.MethodGet
		if strings.HasSuffix(path, "/confirm") {
			method = http.MethodPost
		}
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(method, path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s answered %d unauthenticated, want 401", path, rec.Code)
		}
	}
}

func TestConfirmationSaysSoWhenItIsNotConfigured(t *testing.T) {
	srv := confServer(t, nil)
	rec, _ := confDo(t, srv, http.MethodGet, "/v1/reports/awaiting-confirmation", "")
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("status %d, want 501", rec.Code)
	}
}

// The question has to reach the surface that can answer it. The field kit is
// where a reporter already is, and the block has to say the two things that
// keep it honest: that nothing has been sent, and what the classifier
// thought, so the answer is a correction rather than a guess.
func TestTheFieldKitCarriesTheQuestionBlock(t *testing.T) {
	srv := confServer(t, &fakeConfirmations{})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d for the field kit at /", rec.Code)
	}
	page := rec.Body.String()
	for _, want := range []string{
		"Which problem is this?",
		"/v1/reports/awaiting-confirmation",
		"/confirm",
		// Hard rule 1, on the one surface where a person is about to act.
		"Nothing is sent",
		// The classifier's own guess, so the answer corrects something.
		"It looked like",
		// The platform's words are shown back rather than a cheerful
		// "done", because one of the two endings is "this goes no further".
		"out.note",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the field kit does not carry %q", want)
		}
	}
	// And the block starts hidden: an empty list must not render an empty
	// box that trains a reporter to stop looking at it.
	if !strings.Contains(page, `id="askBox" hidden`) {
		t.Error("the question block is not hidden by default")
	}
}
