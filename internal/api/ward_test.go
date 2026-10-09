package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

type fakeWards struct{ profile store.WardProfile }

func (f *fakeWards) WardProfileFor(_ context.Context, authority, ward string) (store.WardProfile, error) {
	p := f.profile
	p.Authority, p.Ward = authority, ward
	return p, nil
}

func wardServer(t *testing.T, w *fakeWards) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Wards: w, Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func profileWithADesk() store.WardProfile {
	return store.WardProfile{
		Desks: []store.Desk{{
			Category: "waste", Department: "Solid Waste Management — R/Central Ward",
			Officer:     "Assistant Engineer (SWM) R/Central",
			Office:      "Chandavarkar Road, Borivali (West)",
			OfficePhone: "022-28946000", OfficeEmail: "ae01swm.rc@mcgm.gov.in",
			VisitingHours:    "6.30 a.m. – 1.15 p.m.",
			EscalatesTo:      "Assistant Commissioner, R/Central Ward",
			DeadlineHours:    24,
			DeadlineCitation: "R/Central SWM RTI handbook §4(1)(b)(iii)",
			SourceRef:        "https://www.mcgm.gov.in/…/AESWMRCentralWardManuals.pdf",
			RetrievedAt:      time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC),
		}},
		Segments: 207, Metres: 61000,
	}
}

func TestTheWardProfileIsPublic(t *testing.T) {
	rec := httptest.NewRecorder()
	// Deliberately no Authorization header.
	wardServer(t, &fakeWards{profile: profileWithADesk()}).Handler().
		ServeHTTP(rec, httptest.NewRequest("GET", "/v1/public/ward/BMC/R%2FC", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d without a token: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Authority string `json:"authority"`
		Ward      string `json:"ward"`
		Desks     []struct {
			Department       string `json:"department"`
			Phone            string `json:"phone"`
			DeadlineHours    int    `json:"deadline_hours"`
			DeadlineCitation string `json:"deadline_citation"`
			Source           string `json:"source"`
			RetrievedAt      string `json:"retrieved_at"`
		} `json:"desks"`
		Limits []string `json:"limits"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if body.Ward != "R/C" {
		t.Errorf("the ward code did not survive the path: %q", body.Ward)
	}
	if len(body.Desks) != 1 {
		t.Fatalf("expected the desk, got %d", len(body.Desks))
	}
	d := body.Desks[0]
	// Hard rule 2 travels with the fact, not with the page.
	if d.Source == "" || d.RetrievedAt == "" {
		t.Error("a published fact is served without its source or retrieval date")
	}
	if d.DeadlineHours > 0 && d.DeadlineCitation == "" {
		t.Error("a deadline is served without its citation")
	}

	// The limit that matters most: absence of a contract is not absence of
	// an owner.
	var saysWhoElse bool
	for _, l := range body.Limits {
		if strings.Contains(l, "MMRDA") {
			saysWhoElse = true
		}
	}
	if !saysWhoElse {
		t.Error("nothing says a road without a contract may belong to another body")
	}
}

// Nothing a citizen reported may appear here. That is the whole basis for
// serving it without a sign-in — no photograph, no coordinate, no person.
func TestTheWardProfileCarriesNoCitizenData(t *testing.T) {
	rec := httptest.NewRecorder()
	wardServer(t, &fakeWards{profile: profileWithADesk()}).Handler().
		ServeHTTP(rec, httptest.NewRequest("GET", "/v1/public/ward/BMC/R%2FC", nil))

	body := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{
		"report_id", "\"lat\"", "\"lon\"", "accuracy", "account", "media",
		"captured_at", "photograph",
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("the public ward profile leaks %q", forbidden)
		}
	}
}

// Most wards have nothing recorded. That is an ordinary answer and the page
// must be able to say it rather than fail.
func TestAWardWithNoDesksIsNotAnError(t *testing.T) {
	rec := httptest.NewRecorder()
	wardServer(t, &fakeWards{}).Handler().
		ServeHTTP(rec, httptest.NewRequest("GET", "/v1/public/ward/BMC/R%2FN", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"desks":[]`) {
		t.Errorf("an empty ward should serve an empty list: %s", rec.Body.String())
	}
}

// The page itself must be reachable without a sign-in, and must not promise
// anything the platform does not do.
func TestTheWardPageIsPublicAndHonest(t *testing.T) {
	rec := httptest.NewRecorder()
	wardServer(t, &fakeWards{}).Handler().
		ServeHTTP(rec, httptest.NewRequest("GET", "/ward/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d without a token", rec.Code)
	}
	body := rec.Body.String()

	// Hard rule 1: no surface may suggest the platform files anything.
	for _, forbidden := range []string{
		"File complaint", "Submit complaint", "we will file", "File this",
	} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Errorf("the ward page says %q", forbidden)
		}
	}
	// It must scope itself to one body and one ward, not to "your area".
	if !strings.Contains(body, "R/Central") {
		t.Error("the page does not name the ward it describes")
	}
	if !strings.Contains(body, "Brihanmumbai Municipal Corporation") {
		t.Error("the page does not name the body it describes")
	}
	// And it must say what it holds, so nobody mistakes it for a public feed
	// of other people's reports.
	if !strings.Contains(body, "nothing anybody reported") {
		t.Error("the page does not say it carries no citizen data")
	}
}
