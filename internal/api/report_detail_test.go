package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/classify"
)

type fakeDetails struct {
	byID  map[string]ReportDetail
	owner map[string]string
}

func (f *fakeDetails) ReportDetail(_ context.Context, id string) (ReportDetail, bool, error) {
	d, ok := f.byID[id]
	return d, ok, nil
}

func detailServer(t *testing.T, d *fakeDetails) (*Server, string) {
	t.Helper()
	issuer, err := auth.NewIssuer("a-test-signing-secret-32-bytes!!!", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{}, Token: "test-token",
		Account: "field-kit", Issuer: issuer, Details: d,
	})
	if err != nil {
		t.Fatal(err)
	}
	token, err := issuer.Issue("person-1", "p@example.org")
	if err != nil {
		t.Fatal(err)
	}
	return srv, token
}

func sampleDetail() ReportDetail {
	return ReportDetail{
		ID: "rep-1", AccountID: "person-1", Status: "pending",
		Classification: &classify.Decision{
			Result: classify.Result{
				Category: "road_defect", Subcategory: "pothole",
				Severity: "high", Confidence: 0.91,
				Rationale: "Large depression in asphalt",
			},
			Outcome: classify.Accepted,
		},
		Ward: "R/C", Authority: "BMC", WardConfidence: "high",
	}
}

func TestAReportDetailNeedsASignedInCaller(t *testing.T) {
	srv, _ := detailServer(t, &fakeDetails{byID: map[string]ReportDetail{"rep-1": sampleDetail()}})

	req := httptest.NewRequest(http.MethodGet, "/v1/reports/rep-1", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestAPersonSeesTheirOwnReportInFull(t *testing.T) {
	srv, token := detailServer(t, &fakeDetails{byID: map[string]ReportDetail{"rep-1": sampleDetail()}})

	req := httptest.NewRequest(http.MethodGet, "/v1/reports/rep-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	raw := rec.Body.String()
	for _, want := range []string{"pothole", "R/C", "road_defect"} {
		if !strings.Contains(raw, want) {
			t.Errorf("the detail is missing %q: %s", want, raw)
		}
	}
}

func TestAPersonCannotReadSomebodyElsesReport(t *testing.T) {
	// A report carries a precise position and a photograph of a place
	// somebody stood. It is not public just because the id is guessable.
	other := sampleDetail()
	other.AccountID = "someone-else"
	srv, token := detailServer(t, &fakeDetails{byID: map[string]ReportDetail{"rep-1": other}})

	req := httptest.NewRequest(http.MethodGet, "/v1/reports/rep-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	// 404 rather than 403: confirming a report exists tells a stranger
	// something about somebody else's report.
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", rec.Code)
	}
}

func TestAMissingReportIsAPlainNotFound(t *testing.T) {
	srv, token := detailServer(t, &fakeDetails{byID: map[string]ReportDetail{}})

	req := httptest.NewRequest(http.MethodGet, "/v1/reports/nope", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", rec.Code)
	}
}

func TestAnUnenrichedReportStillReturns(t *testing.T) {
	// The capture is the thing that must never be lost. A report whose
	// enrichment has not run yet is still a report, and saying "pending"
	// is better than failing.
	bare := ReportDetail{ID: "rep-1", AccountID: "person-1", Status: "pending"}
	srv, token := detailServer(t, &fakeDetails{byID: map[string]ReportDetail{"rep-1": bare}})

	req := httptest.NewRequest(http.MethodGet, "/v1/reports/rep-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTheFieldKitTokenCannotBrowseReports(t *testing.T) {
	// The shared token names nobody, so it cannot be said to own anything.
	srv, _ := detailServer(t, &fakeDetails{byID: map[string]ReportDetail{"rep-1": sampleDetail()}})

	req := httptest.NewRequest(http.MethodGet, "/v1/reports/rep-1", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Fatal("an anonymous caller must not read a person's report")
	}
}
