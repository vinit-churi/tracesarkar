package api

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/publish"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

type fakePublicIssues struct {
	view  store.PublicIssueView
	found bool
}

func (f *fakePublicIssues) PublicIssue(context.Context, string) (store.PublicIssueView, bool, error) {
	return f.view, f.found, nil
}

type fakeStore struct {
	got     string
	corrupt bool
}

func (f *fakeStore) Get(_ context.Context, key string) ([]byte, error) {
	f.got = key
	if f.corrupt {
		return []byte("truncated upload"), nil
	}
	return tinyJPEG, nil
}

// tinyJPEG is a real, decodable photograph-shaped object, because the public
// image path now reads what it serves.
var tinyJPEG = func() []byte {
	img := image.NewGray(image.Rect(0, 0, 32, 24))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}()

// The exact position of the first sighting, and the account that made it.
const (
	exactLat   = 19.2309441
	exactLon   = 72.8567123
	reporterID = "11f1301d-9ea6-4e54-b4d1-2476768d87b5"
	reference  = "BMC/RC/2026/44871"
)

func anIssue() store.PublicIssueView {
	lodged := time.Date(2026, 10, 9, 9, 10, 0, 0, time.UTC)
	return store.PublicIssueView{
		IssueID:   "3f1a0f9c-0000-4000-8000-000000000001",
		Authority: "BMC", Ward: "R/C",
		Category: "road_defect", Subcategory: "pothole",
		Status: "open", Sightings: 2,
		Lat: exactLat, Lon: exactLon,
		FirstSeen:  time.Date(2026, 10, 5, 18, 22, 0, 0, time.UTC),
		LastSeen:   time.Date(2026, 10, 9, 7, 55, 0, 0, time.UTC),
		NotifiedAt: &lodged,
		Reference:  reference,
		Redaction: &publish.Redaction{
			RanAt:     time.Date(2026, 10, 13, 11, 0, 0, 0, time.UTC),
			Tool:      "pigo-1.4.6/facefinder+edge-density-1",
			OutputKey: "derivative/r1/m1.jpg", Regions: 1, Published: true,
		},
		Events: []store.PublicEvent{
			{Seq: 1, Kind: "reported", At: time.Date(2026, 10, 5, 18, 22, 0, 0, time.UTC),
				Detail: "No open road defect issue lies within 8 m", Hash: "9f2ca41b"},
			{Seq: 2, Kind: "joined", At: time.Date(2026, 10, 9, 7, 55, 0, 0, time.UTC),
				Detail: "A second sighting 4 m away", Hash: "b6132f09"},
		},
	}
}

func publicServer(t *testing.T, f *fakePublicIssues, blobs Blobs) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		PublicIssues: f, Blobs: blobs,
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func getPublicIssue(t *testing.T, f *fakePublicIssues) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	rec := httptest.NewRecorder()
	// Deliberately no Authorization header: this surface is public.
	publicServer(t, f, &fakeStore{}).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/v1/public/issue/"+f.view.IssueID, nil))

	var body map[string]any
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("body is not JSON: %v\n%s", err, rec.Body.String())
		}
	}
	return rec, body
}

func TestThePublicIssueIsServedWithoutASignIn(t *testing.T) {
	rec, _ := getPublicIssue(t, &fakePublicIssues{view: anIssue(), found: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d without a token: %s", rec.Code, rec.Body.String())
	}
}

func TestThePublicIssueDisclosesNothingItShouldNot(t *testing.T) {
	rec, _ := getPublicIssue(t, &fakePublicIssues{view: anIssue(), found: true})
	body := rec.Body.String()

	for _, secret := range []struct{ what, value string }{
		{"the reporter's account", reporterID},
		{"the complaint reference", reference},
		{"the exact latitude", "19.2309"},
		{"the exact longitude", "72.8567"},
		{"the archive key of the original", "media/"},
	} {
		if strings.Contains(body, secret.value) {
			t.Errorf("%s reached a public page: %q appears in the body", secret.what, secret.value)
		}
	}
}

func TestThePublicIssuePositionIsACellAndSaysSo(t *testing.T) {
	_, body := getPublicIssue(t, &fakePublicIssues{view: anIssue(), found: true})

	at, ok := body["at"].(map[string]any)
	if !ok {
		t.Fatalf("no position on the page: %v", body)
	}
	want := publish.Coarsen(exactLat, exactLon)
	if at["lat"] != want.Lat || at["lon"] != want.Lon {
		t.Errorf("position is %v, want the snapped cell %v", at, want)
	}
	// A dot drawn with no stated width implies a precision the page does not
	// have, so the cell size travels with the point.
	if m, _ := at["cell_metres"].(float64); m < 100 {
		t.Errorf("the page does not say how vague its position is: %v", at["cell_metres"])
	}
}

func TestTheTimelineCarriesHashesAndNoActor(t *testing.T) {
	_, body := getPublicIssue(t, &fakePublicIssues{view: anIssue(), found: true})

	events, _ := body["history"].([]any)
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	for _, e := range events {
		m := e.(map[string]any)
		if m["hash"] == "" || m["hash"] == nil {
			t.Error("an event carries no hash; the chain cannot be checked from the page")
		}
		if _, found := m["actor"]; found {
			t.Error("an event carries an actor; that is the reporter's account")
		}
	}
}

func TestCorroborationIsStatedOnlyWhenItIsTrue(t *testing.T) {
	cases := []struct {
		sightings int
		want      bool
	}{{1, false}, {2, true}, {7, true}}
	for _, c := range cases {
		v := anIssue()
		v.Sightings = c.sightings
		_, body := getPublicIssue(t, &fakePublicIssues{view: v, found: true})

		_, found := body["corroboration"]
		if found != c.want {
			t.Errorf("%d sighting(s): corroboration present = %v, want %v", c.sightings, found, c.want)
		}
	}
}

func TestClaimedResolvedAndCitizenConfirmedNeverShareALabel(t *testing.T) {
	labels := map[string]string{}
	for _, status := range []string{"open", "claimed_resolved", "citizen_confirmed", "reopened"} {
		v := anIssue()
		v.Status = status
		_, body := getPublicIssue(t, &fakePublicIssues{view: v, found: true})

		label, _ := body["status_label"].(string)
		if label == "" {
			t.Fatalf("status %q has no label", status)
		}
		if prior, clash := labels[label]; clash {
			t.Errorf("%q and %q share the label %q", prior, status, label)
		}
		labels[label] = status
	}
}

func TestNoImageIsOfferedWithoutAPublishedRedaction(t *testing.T) {
	for _, r := range []*publish.Redaction{
		nil,
		{RanAt: time.Now(), Tool: "t", OutputKey: "derivative/r1/m1.jpg"}, // held
		{RanAt: time.Now(), Tool: "t", Published: true},                   // no output
	} {
		v := anIssue()
		v.Redaction = r
		_, body := getPublicIssue(t, &fakePublicIssues{view: v, found: true})

		if body["image"] != nil {
			t.Errorf("an image was offered on redaction %+v", r)
		}
		if s, _ := body["image_withheld"].(string); s == "" {
			t.Errorf("no image and no explanation, on redaction %+v", r)
		}
	}
}

func TestTheImageEndpointServesOnlyTheRedactedCopy(t *testing.T) {
	blobs := &fakeStore{}
	f := &fakePublicIssues{view: anIssue(), found: true}

	rec := httptest.NewRecorder()
	publicServer(t, f, blobs).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/v1/public/issue/"+f.view.IssueID+"/image", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	if blobs.got != "derivative/r1/m1.jpg" {
		t.Errorf("read %q, want the redacted copy", blobs.got)
	}
	if strings.HasPrefix(blobs.got, "media/") {
		t.Error("the public image endpoint read the unredacted original")
	}
}

func TestTheImageEndpointRefusesAHeldPhotograph(t *testing.T) {
	blobs := &fakeStore{}
	v := anIssue()
	v.Redaction = &publish.Redaction{
		RanAt: time.Now(), Tool: "t", OutputKey: "derivative/r1/m1.jpg",
	}
	f := &fakePublicIssues{view: v, found: true}

	rec := httptest.NewRecorder()
	publicServer(t, f, blobs).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/v1/public/issue/"+v.IssueID+"/image", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404 for a held photograph", rec.Code)
	}
	if blobs.got != "" {
		t.Errorf("the object store was read anyway, for %q", blobs.got)
	}
}

func TestAnUnknownIssueIsNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	publicServer(t, &fakePublicIssues{found: false}, &fakeStore{}).Handler().
		ServeHTTP(rec, httptest.NewRequest("GET", "/v1/public/issue/3f1a0f9c-0000-4000-8000-000000000009", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", rec.Code)
	}
}

func TestThePublicIssueNeverCharacterisesConduct(t *testing.T) {
	_, body := getPublicIssue(t, &fakePublicIssues{view: anIssue(), found: true})
	raw := strings.ToLower(mustJSON(t, body))

	// Hard rule 3. The page states facts and joins; it does not conclude.
	for _, word := range []string{
		"negligen", "corrupt", "fail", "ignored", "scam", "fraud",
		"guilty", "blame", "responsible for the", "should have",
	} {
		if strings.Contains(raw, word) {
			t.Errorf("the page characterises conduct: %q appears in it", word)
		}
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestTheIssuePageIsPublicAndCarriesNoReporterData(t *testing.T) {
	rec := httptest.NewRecorder()
	// Deliberately no Authorization header: this surface is public.
	publicServer(t, &fakePublicIssues{found: true, view: anIssue()}, &fakeStore{}).Handler().
		ServeHTTP(rec, httptest.NewRequest("GET", "/issue/"+anIssue().IssueID+"/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d for the page: %s", rec.Code, rec.Body.String())
	}
	page := rec.Body.String()

	// The page now carries per-issue metadata so that a link unfurls into the
	// facts, which makes what it may NOT carry the thing worth asserting: an
	// unfurl is fetched by a third party and cached out of reach of any later
	// correction.
	for _, leak := range []string{reporterID, reference, "19.2309", "72.8567", "media/"} {
		if strings.Contains(page, leak) {
			t.Errorf("the page carries %q", leak)
		}
	}
	// It has to say the three things the design promises, in the markup, so a
	// reader with no JavaScript is not told something untrue by omission.
	for _, promise := range []string{
		"has been filed with any authority by TraceSarkar",
		"grid cell",
		"obscured",
	} {
		if !strings.Contains(page, promise) {
			t.Errorf("the page does not state %q", promise)
		}
	}
}

// The page is served only after the issue has been read, so a trailing slash
// is the same page and an id the database does not know is not a page at all.
//
// This replaces the earlier rule that every id got identical bytes. That rule
// existed so a crawler could reach nothing before the gate had been consulted;
// the gate is now consulted first instead, which is the same guarantee earlier
// in the request ([D139](../../docs/00-overview/05-decision-log.md)).
func TestTheIssuePageIsServedOnlyAfterTheIssueHasBeenRead(t *testing.T) {
	known := publicServer(t, &fakePublicIssues{found: true, view: anIssue()}, &fakeStore{})
	var first string
	for _, path := range []string{
		"/issue/" + anIssue().IssueID,
		"/issue/" + anIssue().IssueID + "/",
	} {
		rec := httptest.NewRecorder()
		known.Handler().ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: got %d", path, rec.Code)
		}
		if first == "" {
			first = rec.Body.String()
		} else if rec.Body.String() != first {
			t.Errorf("%s served a different page", path)
		}
	}

	unknown := publicServer(t, &fakePublicIssues{found: false}, &fakeStore{})
	for _, path := range []string{
		"/issue/3f1a0f9c-0000-4000-8000-0000000000ff",
		"/issue/not-a-uuid/",
	} {
		rec := httptest.NewRecorder()
		unknown.Handler().ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: got %d, want 404", path, rec.Code)
		}
	}
}

func TestACorruptStoredCopyIsAnErrorRatherThanABrokenImage(t *testing.T) {
	blobs := &fakeStore{corrupt: true}
	f := &fakePublicIssues{view: anIssue(), found: true}

	rec := httptest.NewRecorder()
	publicServer(t, f, blobs).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/v1/public/issue/"+f.view.IssueID+"/image", nil))

	if rec.Code == http.StatusOK {
		t.Fatal("bytes that are not an image were served as a photograph")
	}
}

// A process fact says how the thing photographed is lawfully done. It must
// reach the page in its own block, with its citation, and it must never be
// mistaken for the clock: the Tree Authority's forty-five days are the time to
// decide an application to fell a tree, not the time to answer a report that
// one came down.
func TestPublicIssueStatesTheLawfulProcessApartFromTheClock(t *testing.T) {
	v := anIssue()
	v.Category, v.Subcategory = "environment", "tree felling"
	v.ProcessNotes = []store.ProcessNote{
		{Says: "Felling or transplanting a tree needs the Tree Authority's " +
			"permission, and the Authority has 45 days to decide an application.",
			Citation: "MRTPS Act 2015 notification, Gardens, item 3"},
		{Says: "The published processing fee is ₹6,000 per tree.",
			Citation: "MRTPS Act 2015 notification, Gardens, item 3"},
	}
	// No deadline is attached: nothing binds a report of felling.
	v.DeadlineHours, v.DeadlineCitation = 0, ""

	_, body := getPublicIssue(t, &fakePublicIssues{view: v, found: true})

	process, ok := body["process"].(map[string]any)
	if !ok {
		t.Fatalf("no process block: %v", body["process"])
	}
	facts, _ := process["facts"].([]any)
	if len(facts) != 2 {
		t.Fatalf("got %d process fact(s), want 2", len(facts))
	}
	for i, raw := range facts {
		f, _ := raw.(map[string]any)
		if s, _ := f["says"].(string); strings.TrimSpace(s) == "" {
			t.Errorf("fact %d says nothing", i)
		}
		if c, _ := f["cite"].(string); strings.TrimSpace(c) == "" {
			t.Errorf("fact %d has no citation, so it must not have been shown", i)
		}
	}
	// The disclaimer is the point of the block, not decoration.
	note, _ := process["note"].(string)
	if !strings.Contains(note, "not a deadline") {
		t.Errorf("the block does not say it is not a deadline: %q", note)
	}
	if !strings.Contains(note, "whether anybody followed them") {
		t.Errorf("the block does not disclaim any finding about a party: %q", note)
	}

	// And none of it leaked into the clock.
	lodged, _ := body["lodged"].(map[string]any)
	if _, present := lodged["deadline_says"]; present {
		t.Error("a process fact became a deadline")
	}
}

// Everything else keeps its silence. A pothole has a clock and no process
// block, and a block that appeared on every issue would be noise that trains
// people to skip it.
func TestPublicIssueHasNoProcessBlockWhenNothingIsPublished(t *testing.T) {
	_, body := getPublicIssue(t, &fakePublicIssues{view: anIssue(), found: true})
	if _, present := body["process"]; present {
		t.Errorf("a pothole grew a process block: %v", body["process"])
	}
}
