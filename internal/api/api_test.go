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

// The ward belongs on the capture screen, not in one-time setup. Someone
// surveying Borivali and Dahisar in one afternoon crosses a boundary, and a
// ward set once at sign-in labels every capture after the crossing wrongly —
// which is worse than not recording one, because it looks like evidence.
func TestTheWardCanBeChangedWhileSurveying(t *testing.T) {
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

	// Offered from the wards the platform holds, however it is rendered — the
	// point is that it is not free text with no list behind it.
	if !strings.Contains(body, `list="wards"`) && !strings.Contains(body, `<select id="ward">`) {
		t.Error("the ward is not offered from the wards we hold")
	}
	// Somewhere outside every boundary has to be sayable, and discoverable.
	// Mira Road is a different corporation and calling it a BMC ward poisons
	// the golden set.
	if !strings.Contains(body, "outside") {
		t.Error("there is no way to say you are outside BMC")
	}
	if !strings.Contains(body, "in a BMC ward at all") {
		t.Error("nothing tells the surveyor that option exists")
	}
	// And it must sit with the capture, not behind the sign-in card.
	setup := strings.Index(body, `id="setup"`)
	ward := strings.Index(body, `id="ward"`)
	if setup < 0 || ward < 0 || ward < setup {
		t.Error("the ward chooser is inside the setup card")
	}
}

// The ward is what the resolver computes from the GPS fix, and ward_ground_truth
// exists to score that computation — so it cannot be filled in from the same
// fix, and it must not be guessed. Almost nobody standing on a street in Mumbai
// knows their lettered ward; what they do know is the street.
func TestTheFieldKitAsksForALandmarkAndNotAGuessedWard(t *testing.T) {
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

	if !strings.Contains(body, `id="where"`) {
		t.Error("it does not ask where the surveyor is")
	}
	if !strings.Contains(body, "landmark") {
		t.Error("the landmark is not sent with the capture")
	}
	// Unknown must be the default, so an unknown is recorded as unknown. The
	// field starts empty and says so, rather than pre-selecting a ward the
	// surveyor would then have to notice and correct.
	if !strings.Contains(body, `placeholder="not sure"`) {
		t.Error("the ward field does not default to not knowing")
	}
	// The first fix removed one default and left another in the settings
	// getter, so a capture taken past Dahisar went in as R/C ground truth.
	// No ward may be filled in from anywhere but the surveyor's own choice.
	if strings.Contains(body, `|| 'R/C'`) {
		t.Error("the ward still defaults to a guess")
	}
	// Nor may it carry over from the last capture: a walk crosses a ward
	// boundary without the surveyor noticing, and a remembered ward labels
	// every capture after the crossing wrongly.
	if strings.Contains(body, "ts_ward") {
		t.Error("the ward is remembered between captures")
	}
	if !strings.Contains(body, "els('ward').value = ''") {
		t.Error("the ward is not cleared after a capture is saved")
	}
}

// Six categories and forty subcategories in a dropdown is a scroll on a phone,
// one-handed, in the sun. Typed and filtered is faster, and the list is still
// the taxonomy's — a free-text box that accepts anything would send labels the
// server drops.
func TestTheFieldKitFiltersRatherThanScrolls(t *testing.T) {
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

	for _, id := range []string{"label", "ward"} {
		if !strings.Contains(body, `<input id="`+id+`" list=`) {
			t.Errorf("%q is not an autocomplete", id)
		}
		if strings.Contains(body, `<select id="`+id+`">`) {
			t.Errorf("%q is still a dropdown", id)
		}
	}
	// Typed freely means it has to be checked before it is sent, or a
	// photograph arrives with a label the server silently drops.
	if !strings.Contains(body, "KNOWN_LABELS") {
		t.Error("a typed label is never checked against the taxonomy")
	}
	if !strings.Contains(body, "will not be recorded") {
		t.Error("nothing tells the surveyor their label is unusable")
	}
}
