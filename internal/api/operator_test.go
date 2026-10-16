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

	"github.com/vinit-churi/tracesarkar/internal/store"
)

type fakeOperator struct {
	health   store.Health
	err      error
	runs     int
	trend    store.Trend
	trendErr error
	days     int
}

func (f *fakeOperator) OperatorHealth(_ context.Context, runLimit int) (store.Health, error) {
	f.runs = runLimit
	return f.health, f.err
}

func (f *fakeOperator) OperatorTrend(_ context.Context, days int) (store.Trend, error) {
	f.days = days
	if f.trend.Days == 0 {
		return store.Trend{}, f.trendErr
	}
	return f.trend, f.trendErr
}

func aHealth() store.Health {
	finished := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	return store.Health{
		At:       time.Date(2026, 10, 16, 10, 0, 0, 0, time.UTC),
		Captures: 5,
		Queues: []store.Queue{
			{Name: "captures not classified", Count: 0, Clears: "the sweep's classify stage"},
			{Name: "captures waiting for the citizen to say which problem this is",
				Count: 3, Clears: "the person who sent it, and no job can do it for them"},
		},
		Issues: []store.IssueCount{
			{Status: "open", Count: 2}, {Status: "claimed_resolved", Count: 1},
			{Status: "citizen_confirmed", Count: 0}, {Status: "reopened", Count: 0},
		},
		Sources: []store.SnapshotStatus{
			{SourceID: "bmc_roads", Endpoint: "works", Records: 2237, Attempts: 9, Failures: 1},
		},
		Runs: []store.RunRow{
			{Command: "sweep", StartedAt: finished, FinishedAt: &finished, OK: true},
			{Command: "run", StartedAt: finished, OK: false}, // never returned
			{Command: "watch", StartedAt: finished, FinishedAt: &finished, OK: false,
				Error: "upstream returned 503"},
		},
	}
}

func opsServer(t *testing.T, op Operator) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Operator: op, Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func opsGet(t *testing.T, srv *Server, path, token string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	var body map[string]any
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("body is not JSON: %v\n%s", err, rec.Body.String())
		}
	}
	return rec, body
}

// A count of what is stuck is an operational signal, and the ratio's public
// version belongs on a surface built for publishing it.
func TestOperatorHealthIsNotPublic(t *testing.T) {
	srv := opsServer(t, &fakeOperator{health: aHealth()})
	if rec, _ := opsGet(t, srv, "/v1/ops/health", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated request answered %d, want 401", rec.Code)
	}
}

func TestOperatorHealthReportsEveryQueueWithWhatClearsIt(t *testing.T) {
	op := &fakeOperator{health: aHealth()}
	rec, body := opsGet(t, opsServer(t, op), "/v1/ops/health", "test-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	queues, ok := body["queues"].([]any)
	if !ok || len(queues) != 2 {
		t.Fatalf("queues is %v", body["queues"])
	}
	for _, q := range queues {
		m := q.(map[string]any)
		if m["clears"] == nil || m["clears"] == "" {
			t.Errorf("queue %v does not say what clears it", m["name"])
		}
	}

	// The two counts, never a percentage.
	if body["claimed_resolved"] != float64(1) || body["citizen_confirmed"] != float64(0) {
		t.Errorf("ratio is %v / %v", body["claimed_resolved"], body["citizen_confirmed"])
	}
	for _, forbidden := range []string{"percent", "ratio\":", "%"} {
		if strings.Contains(rec.Body.String(), forbidden) {
			t.Errorf("the response carries %q, which a small-numbers percentage must not be",
				forbidden)
		}
	}

	// Every status, including the zeros.
	issues, _ := body["issues"].([]any)
	if len(issues) != 4 {
		t.Errorf("issues reports %d statuses, want all 4", len(issues))
	}
}

// A run that never returned is a different failure from one that returned an
// error, and the two must not read the same.
func TestARunThatNeverReturnedIsNotARunThatFailed(t *testing.T) {
	_, body := opsGet(t, opsServer(t, &fakeOperator{health: aHealth()}),
		"/v1/ops/health", "test-token")
	runs, _ := body["runs"].([]any)
	if len(runs) != 3 {
		t.Fatalf("runs is %v", body["runs"])
	}
	var neverReturned, failed int
	for _, r := range runs {
		m := r.(map[string]any)
		if m["never_returned"] == true {
			neverReturned++
			if m["finished_at"] != nil {
				t.Error("a run marked never_returned has a finish time")
			}
		}
		if m["error"] != nil {
			failed++
		}
	}
	if neverReturned != 1 {
		t.Errorf("%d runs marked as never returned, want 1", neverReturned)
	}
	if failed != 1 {
		t.Errorf("%d runs carry an error, want 1", failed)
	}
}

func TestOperatorHealthValidatesTheRunsParameter(t *testing.T) {
	op := &fakeOperator{health: aHealth()}
	srv := opsServer(t, op)
	for _, bad := range []string{"-1", "101", "lots", "1.5"} {
		if rec, _ := opsGet(t, srv, "/v1/ops/health?runs="+bad, "test-token"); rec.Code != http.StatusBadRequest {
			t.Errorf("runs=%s answered %d, want 400", bad, rec.Code)
		}
	}
	if rec, _ := opsGet(t, srv, "/v1/ops/health?runs=0", "test-token"); rec.Code != http.StatusOK {
		t.Errorf("runs=0 answered %d", rec.Code)
	}
	if op.runs != 0 {
		t.Errorf("runs=0 passed %d to the store", op.runs)
	}
}

// A dashboard that silently shows zero is worse than one that says it could
// not count.
func TestOperatorHealthSaysSoWhenItCannotCount(t *testing.T) {
	rec, _ := opsGet(t, opsServer(t, &fakeOperator{err: errors.New("pool exhausted")}),
		"/v1/ops/health", "test-token")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "pool exhausted") {
		t.Errorf("the response leaks the internal error: %s", rec.Body.String())
	}
}

// 501 rather than a stack trace, matching every other optional surface.
func TestOperatorHealthSaysSoWhenItIsNotConfigured(t *testing.T) {
	rec, _ := opsGet(t, opsServer(t, nil), "/v1/ops/health", "test-token")
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("status %d, want 501", rec.Code)
	}
}

// The console is a page, and it must not be a moderation surface: the places
// that take a human verdict are /review/, /label/ and /truth/.
func TestTheConsolePageIsServedAndDoesNotModerate(t *testing.T) {
	srv := opsServer(t, &fakeOperator{health: aHealth()})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ops/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d for /ops/", rec.Code)
	}
	page := rec.Body.String()
	for _, want := range []string{"Collection health", "What clears it",
		"the authority recorded as attended", "a resident photographed fixed"} {
		if !strings.Contains(page, want) {
			t.Errorf("the console does not carry %q", want)
		}
	}
	// No verdict, no publish, no delete.
	for _, forbidden := range []string{"method: 'POST'", `method:"POST"`, "verdict", "publish", "delete"} {
		if strings.Contains(page, forbidden) && !strings.Contains(page, "/v1/auth/login") {
			t.Errorf("the console appears to write: %q", forbidden)
		}
	}
}

func aTrend() store.Trend {
	return store.Trend{
		Days: 3,
		Fetches: []store.FetchDay{
			{Day: "2026-10-14", Attempts: 9, Failures: 0},
			{Day: "2026-10-15", Attempts: 0, Failures: 0},
			{Day: "2026-10-16", Attempts: 9, Failures: 2},
		},
		Runs: []store.RunDay{
			{Day: "2026-10-14", OK: 4},
			{Day: "2026-10-15", OK: 0},
			{Day: "2026-10-16", OK: 3, Failed: 1, NeverReturned: 1},
		},
		Captures: []store.CaptureDay{
			{Day: "2026-10-14", Captures: 2, Issues: 1},
			{Day: "2026-10-15"},
			{Day: "2026-10-16", Captures: 1},
		},
	}
}

func TestTheTrendKeepsADayOnWhichNothingHappened(t *testing.T) {
	op := &fakeOperator{trend: aTrend()}
	rec, body := opsGet(t, opsServer(t, op), "/v1/ops/trend?days=3", "test-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	fetches, _ := body["fetches"].([]any)
	if len(fetches) != 3 {
		t.Fatalf("fetches has %d buckets, want 3", len(fetches))
	}
	// The middle day is the one that matters: collection did not run, and a
	// series that omitted it would draw that day as if it never existed.
	quiet := fetches[1].(map[string]any)
	if quiet["day"] != "2026-10-15" || quiet["attempts"] != float64(0) {
		t.Errorf("the quiet day is %v", quiet)
	}
	if op.days != 3 {
		t.Errorf("the handler asked for %d days", op.days)
	}

	// The three run outcomes stay separate.
	runs, _ := body["runs"].([]any)
	last := runs[2].(map[string]any)
	if last["failed"] != float64(1) || last["never_returned"] != float64(1) {
		t.Errorf("the run outcomes were merged: %v", last)
	}

	// And the gap is named rather than left for somebody to discover.
	note, _ := body["no_history_for"].(string)
	if !strings.Contains(note, "queue depth") {
		t.Errorf("the response does not name what it has no history for: %q", note)
	}
}

func TestTheTrendValidatesItsWindow(t *testing.T) {
	srv := opsServer(t, &fakeOperator{trend: aTrend()})
	for _, bad := range []string{"0", "-7", "181", "a%20month", "7.5"} {
		if rec, _ := opsGet(t, srv, "/v1/ops/trend?days="+bad, "test-token"); rec.Code != http.StatusBadRequest {
			t.Errorf("days=%s answered %d, want 400", bad, rec.Code)
		}
	}
}

func TestTheTrendIsNotPublic(t *testing.T) {
	srv := opsServer(t, &fakeOperator{trend: aTrend()})
	if rec, _ := opsGet(t, srv, "/v1/ops/trend", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated request answered %d, want 401", rec.Code)
	}
}

// The chart is drawn in the page, so what it promises has to be in the page.
func TestTheConsoleDrawsTheTrendWithoutALibrary(t *testing.T) {
	srv := opsServer(t, &fakeOperator{health: aHealth(), trend: aTrend()})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ops/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	page := rec.Body.String()

	for _, want := range []string{
		"Is this getting worse",
		"/v1/ops/trend?days=",
		"Fetches per day",
		"Collector runs per day",
		// The three run outcomes stay separate in the legend too.
		"never returned",
		// A quiet day is drawn as a hairline, not omitted. Asserted on the
		// class and the tooltip the browser actually renders, because the
		// first version of this test matched the word "nothing" in a code
		// comment and would have passed with the branch deleted.
		`'<i class="none"></i></div>'`,
		`+ ': nothing">'`,
		// And the gap is printed where somebody will read it.
		"No history for",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the console does not carry %q", want)
		}
	}
	// No chart library, from anywhere. This page is embedded in the binary
	// and loads nothing over the network.
	for _, forbidden := range []string{"<script src", "cdn.", "unpkg", "chart.js", "d3."} {
		if strings.Contains(page, forbidden) {
			t.Errorf("the console loads something external: %q", forbidden)
		}
	}
}
