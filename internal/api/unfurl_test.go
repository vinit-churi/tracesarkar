package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/share"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

func getPage(t *testing.T, f *fakePublicIssues, id string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	shareServer(t, f, &fakeHandles{}).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/issue/"+id, nil))
	return rec
}

// A bare link in a chat has to become the argument, or the share kit stops at
// the person who was already looking at the page.
func TestALinkUnfurlsIntoTheFactsRatherThanTheAppName(t *testing.T) {
	f := &fakePublicIssues{view: anIssue(), found: true}
	rec := getPage(t, f, f.view.IssueID)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	html := rec.Body.String()

	for _, want := range []string{
		`property="og:title"`,
		`property="og:description"`,
		`property="og:url"`,
		`name="twitter:card"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("the page carries no %s", want)
		}
	}
	if !strings.Contains(html, "R/C ward") {
		t.Error("the unfurl does not say which ward")
	}
	if !strings.Contains(html, "Pothole") {
		t.Error("the unfurl does not say what the problem is")
	}
}

// Hard rule 5, at the one place it is easiest to lose: an unfurl image is
// fetched by a third party and cached on their servers.
func TestNoUnfurlImageExistsForAPhotographTheGateWithheld(t *testing.T) {
	v := anIssue()
	v.Redaction.Published = false
	rec := getPage(t, &fakePublicIssues{view: v, found: true}, v.IssueID)
	html := rec.Body.String()

	if strings.Contains(html, "og:image") {
		t.Error("an unfurl image was offered for a photograph the gate held")
	}
	// Without an image the large card would render a blank panel, so the
	// smaller card is correct rather than cosmetic.
	if !strings.Contains(html, `content="summary"`) {
		t.Error("the card type was not stepped down when there is no image")
	}

	ok := anIssue()
	rec = getPage(t, &fakePublicIssues{view: ok, found: true}, ok.IssueID)
	html = rec.Body.String()
	if !strings.Contains(html, "og:image") {
		t.Error("a published photograph produced no unfurl image")
	}
	if !strings.Contains(html, "/card") {
		t.Error("the unfurl image is not the annotated card")
	}
	if !strings.Contains(html, `content="summary_large_image"`) {
		t.Error("a published photograph should get the large card")
	}
}

// D130 said the shell is the same bytes for every id so that a crawler reaches
// nothing before the gate has been consulted. Per-issue metadata means the
// gate is consulted earlier instead — which only holds if an unknown id is
// refused here rather than by the page's own JavaScript.
func TestACrawlerGetsNothingForAnIssueThatDoesNotExist(t *testing.T) {
	rec := getPage(t, &fakePublicIssues{found: false}, "3f1a0f9c-0000-4000-8000-0000000000ff")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "og:image") {
		t.Error("an unfurl image was offered for an issue that does not exist")
	}
}

// Hard rule 4 and the reference rule, restated for the one copy of the page a
// third party keeps.
func TestTheUnfurlCarriesNoExactPositionAndNoReference(t *testing.T) {
	rec := getPage(t, &fakePublicIssues{view: anIssue(), found: true}, anIssue().IssueID)
	html := rec.Body.String()
	for _, secret := range []string{"19.2309441", "72.8567123", reference, reporterID} {
		if strings.Contains(html, secret) {
			t.Errorf("the page leaked %q into its metadata", secret)
		}
	}
}

// Everything injected comes from a database this process does not control the
// contents of. An unescaped quote would end the attribute and the rest would
// be markup.
func TestInjectedMetadataIsEscaped(t *testing.T) {
	v := anIssue()
	v.Subcategory = `pothole" /><script>alert(1)</script><meta x="`
	rec := getPage(t, &fakePublicIssues{view: v, found: true}, v.IssueID)
	html := rec.Body.String()
	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Errorf("markup from the database reached the page:\n%s", html[:600])
	}
}

// Hard rule 3 reaches the unfurl too: the description is the sentence most
// people will read, and it is the one they cannot click through to correct.
func TestTheUnfurlDescriptionAssertsNoWrongdoing(t *testing.T) {
	for _, st := range []string{"open", "claimed_resolved", "citizen_confirmed", "reopened"} {
		v := anIssue()
		v.Status = st
		rec := getPage(t, &fakePublicIssues{view: v, found: true}, v.IssueID)
		d := metaContent(rec.Body.String(), `property="og:description"`)
		if d == "" {
			t.Errorf("%s: no description", st)
			continue
		}
		if vs := share.Check(d); len(vs) != 0 {
			t.Errorf("%s: the unfurl description fails the tone check %+v: %s", st, vs, d)
		}
	}
}

// claimed_resolved and citizen_confirmed never share a label (G14), including
// in the one line most readers will ever see.
func TestTheUnfurlDistinguishesWhoSaidItWasFixed(t *testing.T) {
	claimed, citizen := anIssue(), anIssue()
	claimed.Status, citizen.Status = "claimed_resolved", "citizen_confirmed"
	a := metaContent(getPage(t, &fakePublicIssues{view: claimed, found: true}, claimed.IssueID).Body.String(),
		`property="og:description"`)
	b := metaContent(getPage(t, &fakePublicIssues{view: citizen, found: true}, citizen.IssueID).Body.String(),
		`property="og:description"`)
	if a == b {
		t.Errorf("both statuses unfurl as %q", a)
	}
}

// Without a database the page must still serve, as every other optional
// surface does — the shell alone is a working page.
func TestThePageStillServesWhenPublicIssuesIsNotConfigured(t *testing.T) {
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest("GET",
		"/issue/3f1a0f9c-0000-4000-8000-000000000001", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<title>") {
		t.Error("the shell did not serve")
	}
}

// The page is deliberately not indexed. An unfurl crawler reads the metadata
// it is sent and does not need the page in a search index, so this stays.
func TestTheIssuePageIsStillNotIndexed(t *testing.T) {
	rec := getPage(t, &fakePublicIssues{view: anIssue(), found: true}, anIssue().IssueID)
	if !strings.Contains(rec.Body.String(), `name="robots" content="noindex"`) {
		t.Error("the issue page lost its noindex")
	}
}

// The status labels are written for a chip, with no full stop. Joined into a
// paragraph without one they run together: "Open Nobody has recorded telling
// the authority yet".
func TestTheUnfurlDescriptionReadsAsSentences(t *testing.T) {
	for _, st := range []string{"open", "claimed_resolved", "citizen_confirmed", "reopened"} {
		v := anIssue()
		v.Status = st
		d := metaContent(getPage(t, &fakePublicIssues{view: v, found: true}, v.IssueID).Body.String(),
			`property="og:description"`)
		first := strings.SplitN(d, ".", 2)
		if len(first) != 2 || strings.TrimSpace(first[1]) == "" {
			t.Errorf("%s: description is one run-on: %q", st, d)
			continue
		}
		// The sentence after the status must start with a capital, which it
		// cannot if the status ran into it.
		rest := strings.TrimSpace(first[1])
		if r := rune(rest[0]); r >= 'a' && r <= 'z' {
			t.Errorf("%s: sentences run together: %q", st, d)
		}
	}
}

func metaContent(html, marker string) string {
	i := strings.Index(html, marker)
	if i < 0 {
		return ""
	}
	rest := html[i:]
	j := strings.Index(rest, `content="`)
	if j < 0 {
		return ""
	}
	rest = rest[j+len(`content="`):]
	k := strings.Index(rest, `"`)
	if k < 0 {
		return ""
	}
	return rest[:k]
}

var _ = store.PublicIssueView{}
