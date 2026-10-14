package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/redact"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

type fakeTruth struct {
	queue    []store.PendingTruth
	refs     map[string]store.MediaRef // mediaID -> row
	saved    []store.RedactionTruth
	err      error
	askedFor string
}

func (f *fakeTruth) MediaAwaitingTruth(_ context.Context, account string, _ int) ([]store.PendingTruth, error) {
	f.askedFor = account
	return f.queue, f.err
}

func (f *fakeTruth) SaveRedactionTruth(_ context.Context, in store.RedactionTruth) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, in)
	return nil
}

func (f *fakeTruth) MediaByID(_ context.Context, mediaID string) (store.MediaRef, error) {
	ref, ok := f.refs[mediaID]
	if !ok {
		return store.MediaRef{}, store.ErrNotFound
	}
	return ref, nil
}

func truthServer(t *testing.T, truth *fakeTruth, blobs *fakeBlobs) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{},
		Media:   blobs,
		Truth:   truth,
		Blobs:   blobs,
		Token:   "test-token",
		Account: "account-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func mine() *fakeTruth {
	return &fakeTruth{
		queue: []store.PendingTruth{{
			MediaID: "m1", ReportID: "r1", ArchiveKey: "media/r1/a.jpg",
			ContentType: "image/jpeg", CapturedAt: time.Now().UTC(),
		}},
		refs: map[string]store.MediaRef{
			"m1": {Key: "media/r1/a.jpg", ContentType: "image/jpeg", Owner: "account-1", ReportID: "r1"},
			"m2": {Key: "media/r9/a.jpg", ContentType: "image/jpeg", Owner: "someone-else", ReportID: "r9"},
		},
	}
}

func TestTheAnnotationQueueIsScopedToTheSignedInAccount(t *testing.T) {
	truth := mine()
	rec := httptest.NewRecorder()
	truthServer(t, truth, &fakeBlobs{}).Handler().
		ServeHTTP(rec, authed("GET", "/v1/redaction/truth/queue", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	if truth.askedFor != "account-1" {
		t.Errorf("the queue was read for %q, not the requester", truth.askedFor)
	}

	var body struct {
		Items []struct {
			MediaID  string `json:"media_id"`
			MediaURL string `json:"media_url"`
		} `json:"items"`
		Conditions []string `json:"conditions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 1 || body.Items[0].MediaID != "m1" {
		t.Fatalf("items came back as %+v", body.Items)
	}
	// The page fetches the photograph through our own endpoint. The storage
	// key is internal and would let a page read the private bucket directly.
	if strings.Contains(rec.Body.String(), "media/r1/a.jpg") {
		t.Errorf("the queue exposes the storage key: %s", rec.Body.String())
	}
	if !strings.Contains(body.Items[0].MediaURL, "m1") {
		t.Errorf("media url %q does not name the photograph to annotate", body.Items[0].MediaURL)
	}
	// The vocabulary is served rather than copied into the page, so a
	// condition the breakdown does not know cannot be offered.
	if len(body.Conditions) != len(redact.Conditions()) {
		t.Errorf("conditions came back as %v", body.Conditions)
	}
}

func TestAnAnnotationIsRecordedAgainstThePhotographTheServerLookedUp(t *testing.T) {
	truth := mine()
	rec := httptest.NewRecorder()
	truthServer(t, truth, &fakeBlobs{}).Handler().ServeHTTP(rec, authed("POST",
		"/v1/redaction/truth/m1", `{
			"boxes": [{"kind": "face", "min_x": 0.1, "min_y": 0.1, "max_x": 0.2, "max_y": 0.3}],
			"people_present": true,
			"conditions": ["night", "distance"],
			"notes": "two pedestrians",
			"report_id": "r-somebody-elses"
		}`))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	if len(truth.saved) != 1 {
		t.Fatalf("saved %d judgements", len(truth.saved))
	}
	got := truth.saved[0]
	// The report is read from the photograph's own row. A report id taken from
	// the request body would let one account's judgement be filed against
	// another account's capture.
	if got.ReportID != "r1" || got.MediaID != "m1" {
		t.Errorf("recorded against %s / %s", got.ReportID, got.MediaID)
	}
	if got.AnnotatedBy != "account-1" {
		t.Errorf("annotated by %q, want the requester", got.AnnotatedBy)
	}
	if len(got.Boxes) != 1 || got.Boxes[0].Kind != redact.Face {
		t.Errorf("boxes recorded as %+v", got.Boxes)
	}
	if !got.PeoplePresent || len(got.Conditions) != 2 {
		t.Errorf("judgement recorded as %+v", got)
	}
}

func TestNothingHereIsAJudgementAndIsRecorded(t *testing.T) {
	// A photograph with no face and no plate in it is the other half of the
	// set: it is the only thing that can measure a false positive, and it must
	// be distinguishable from a photograph nobody has looked at.
	truth := mine()
	rec := httptest.NewRecorder()
	truthServer(t, truth, &fakeBlobs{}).Handler().ServeHTTP(rec, authed("POST",
		"/v1/redaction/truth/m1", `{"boxes": [], "people_present": false, "conditions": ["daylight"]}`))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	if len(truth.saved) != 1 || len(truth.saved[0].Boxes) != 0 {
		t.Fatalf("saved %+v", truth.saved)
	}
}

func TestAnAnnotationOnSomebodyElsesPhotographIsNotFound(t *testing.T) {
	truth := mine()
	for _, path := range []string{"/v1/redaction/truth/m2", "/v1/redaction/truth/nope"} {
		rec := httptest.NewRecorder()
		truthServer(t, truth, &fakeBlobs{}).Handler().ServeHTTP(rec, authed("POST",
			path, `{"boxes": [], "people_present": false}`))

		// Not 403: telling the caller the photograph exists but is not theirs
		// discloses that it exists.
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s got %d, want 404: %s", path, rec.Code, rec.Body.String())
		}
	}
	if len(truth.saved) != 0 {
		t.Errorf("saved a judgement about a photograph that is not the requester's: %+v", truth.saved)
	}
}

func TestAnAnnotationThatCannotBeScoredIsRefusedWithAReason(t *testing.T) {
	cases := []struct {
		name string
		body string
		says string
	}{
		{
			name: "a box past the edge of the photograph",
			body: `{"boxes": [{"kind": "face", "min_x": 0.5, "min_y": 0.5, "max_x": 1.6, "max_y": 0.9}], "people_present": true}`,
			says: "fraction",
		},
		{
			name: "a kind redaction does not look for",
			body: `{"boxes": [{"kind": "tattoo", "min_x": 0.1, "min_y": 0.1, "max_x": 0.2, "max_y": 0.2}], "people_present": true}`,
			says: "tattoo",
		},
		{
			name: "a condition the breakdown does not know",
			body: `{"boxes": [], "people_present": false, "conditions": ["nite"]}`,
			says: "nite",
		},
		{
			name: "a face with nobody present",
			body: `{"boxes": [{"kind": "face", "min_x": 0.1, "min_y": 0.1, "max_x": 0.2, "max_y": 0.2}], "people_present": false}`,
			says: "nobody",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			truth := mine()
			rec := httptest.NewRecorder()
			truthServer(t, truth, &fakeBlobs{}).Handler().
				ServeHTTP(rec, authed("POST", "/v1/redaction/truth/m1", c.body))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got %d, want 400: %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(strings.ToLower(rec.Body.String()), c.says) {
				t.Errorf("the refusal does not say why: %s", rec.Body.String())
			}
			if len(truth.saved) != 0 {
				t.Errorf("stored it anyway: %+v", truth.saved)
			}
		})
	}
}

func TestThePhotographToAnnotateIsServedOnlyToItsOwner(t *testing.T) {
	truth := mine()
	blobs := &fakeBlobs{put: map[string][]byte{
		"media/r1/a.jpg": []byte("not really a jpeg"),
		"media/r9/a.jpg": []byte("somebody else's"),
	}}
	srv := truthServer(t, truth, blobs)

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, authed("GET", "/v1/media/m1", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("own photograph got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "not really a jpeg" {
		t.Errorf("served %q", rec.Body.String())
	}
	// The original is what the detector reads and what a face is still in. A
	// shared cache must not keep a copy of it.
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "private") {
		t.Errorf("Cache-Control = %q, want private", cc)
	}

	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, authed("GET", "/v1/media/m2", ""))
	if rec.Code != http.StatusNotFound {
		t.Errorf("another account's photograph got %d, want 404", rec.Code)
	}
}

func TestTheAnnotationSurfaceIsAbsentRatherThanBrokenWhenUnconfigured(t *testing.T) {
	srv, err := New(Options{Reports: &fakeReports{}, Media: &fakeBlobs{},
		Token: "test-token", Account: "account-1"})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []*http.Request{
		authed("GET", "/v1/redaction/truth/queue", ""),
		authed("POST", "/v1/redaction/truth/m1", `{"boxes": []}`),
		authed("GET", "/v1/media/m1", ""),
	} {
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, r)
		if rec.Code != http.StatusNotImplemented {
			t.Errorf("%s %s got %d, want 501", r.Method, r.URL.Path, rec.Code)
		}
	}
}
