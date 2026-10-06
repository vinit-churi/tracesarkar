package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The methodology page is the one surface that must work without signing in.
// A platform that publishes facts about named parties has to publish how it got
// them, and a page nobody can reach proves nothing.
func TestMethodologyIsPublic(t *testing.T) {
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	// Deliberately no Authorization header.
	srv.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/methodology/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d without a token, want 200", rec.Code)
	}
	body := rec.Body.String()

	// The limitations are the part that makes the rest worth reading.
	for _, must := range []string{
		"Limitations",
		"Classification accuracy is unmeasured",
		"2025:BHC-OS:18736-DB",
		"Nothing is ever filed automatically",
	} {
		if !strings.Contains(body, must) {
			t.Errorf("the page does not say %q", must)
		}
	}
	// It must not claim a number the project has not measured. Checked past
	// the stylesheet, because a table is allowed to be 100% wide.
	prose := body
	if i := strings.Index(body, "</style>"); i >= 0 {
		prose = body[i:]
	}
	if strings.Contains(prose, "100%") {
		t.Error("the page claims 100% in its prose; the honest floor is 92.9%")
	}
	if !strings.Contains(prose, "92.9%") {
		t.Error("the page does not show the measured precision floor")
	}
}

// The field kit files captures against whoever made them. It used to take the
// shared API token, which files them against nobody: they never appear in that
// person's own reports, and the labelling queue is scoped to the account that
// made the capture, so a day of fieldwork filed that way is a day of
// photographs nobody can label.
func TestFieldKitSignsInRatherThanTakingTheSharedToken(t *testing.T) {
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
	body := rec.Body.String()

	if !strings.Contains(body, "/v1/auth/login") {
		t.Error("the field kit does not sign in")
	}
	if strings.Contains(body, "the API_TOKEN the server was started with") {
		t.Error("the field kit still asks for the shared token")
	}
	// The ward it defaults to has to be the one being surveyed. R/S was the
	// pilot ward until 29 September, when it became R/C.
	if strings.Contains(body, `value="R/S"`) {
		t.Error("the field kit still defaults to the old pilot ward")
	}
}

// The field kit must not keep its own copy of the vocabulary. It had one, and
// it had drifted: missing_manhole_cover, garbage, not_civic — none of them
// values the taxonomy holds, so every photograph labelled through it would
// have been unscoreable.
func TestFieldKitDoesNotCarryItsOwnTaxonomy(t *testing.T) {
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	body := rec.Body.String()

	// Checked as markup rather than as text: the comment explaining why these
	// were removed necessarily names them.
	for _, stale := range []string{
		"missing_manhole_cover", "utility_dig_damage", "faded_markings",
		"garbage", "not_civic", "blocked_drain", "streetlight_out",
	} {
		for _, form := range []string{">" + stale + "<", `value="` + stale + `"`} {
			if strings.Contains(body, form) {
				t.Errorf("the field kit still offers %q", stale)
			}
		}
	}
	if !strings.Contains(body, "/v1/taxonomy") {
		t.Error("the field kit does not fetch the taxonomy")
	}
}
