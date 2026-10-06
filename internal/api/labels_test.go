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

type fakeLabels struct {
	queue  []store.PendingLabel
	saved  []store.ReportLabel
	ownsBy map[string]string // reportID -> accountID
	keyBy  map[string]string // reportID -> archive key
	err    error
	// askedFor records which account the queue was read for, so the scoping
	// is asserted rather than assumed.
	askedFor string
}

func (f *fakeLabels) UnlabelledReports(_ context.Context, account string, _ int) ([]store.PendingLabel, error) {
	f.askedFor = account
	return f.queue, f.err
}
func (f *fakeLabels) SaveReportLabel(_ context.Context, in store.ReportLabel) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, in)
	return nil
}
func (f *fakeLabels) MediaFor(_ context.Context, reportID string) (key, contentType, owner string, err error) {
	k, ok := f.keyBy[reportID]
	if !ok {
		return "", "", "", store.ErrNotFound
	}
	return k, "image/jpeg", f.ownsBy[reportID], nil
}

func (f *fakeBlobs) Get(_ context.Context, key string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	b, ok := f.put[key]
	if !ok {
		return nil, store.ErrNotFound
	}
	return b, nil
}

func labelServer(t *testing.T, labels *fakeLabels, blobs *fakeBlobs) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{},
		Media:   blobs,
		Labels:  labels,
		Blobs:   blobs,
		Token:   "test-token",
		Account: "account-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func authed(method, path, body string) *http.Request {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer test-token")
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	return r
}

func TestLabelQueueDoesNotRevealWhatTheModelSaid(t *testing.T) {
	labels := &fakeLabels{queue: []store.PendingLabel{
		{ReportID: "r1", ArchiveKey: "media/r1/a.jpg", ContentType: "image/jpeg",
			Ward: "R/C", CapturedAt: time.Now().UTC()},
	}}
	rec := httptest.NewRecorder()
	labelServer(t, labels, &fakeBlobs{}).Handler().ServeHTTP(rec, authed("GET", "/v1/label/queue", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	// Showing the classifier's answer turns labelling into grading, and the
	// eval set stops being independent of the thing it measures.
	for _, leak := range []string{"category", "subcategory", "confidence", "outcome"} {
		if strings.Contains(strings.ToLower(body), leak) {
			t.Errorf("the queue leaks %q to the labeller: %s", leak, body)
		}
	}
	// The archive key is internal; the page fetches through our own endpoint.
	if strings.Contains(body, "media/r1/a.jpg") {
		t.Errorf("the queue exposes the storage key: %s", body)
	}
	if labels.askedFor != "account-1" {
		t.Errorf("the queue was read for %q, not the requester", labels.askedFor)
	}
}

func TestSavingALabel(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode int
		wantSub  string
	}{
		{
			name:     "a subcategory from the taxonomy is accepted",
			body:     `{"label":"pothole","frame_type":"close"}`,
			wantCode: http.StatusOK,
			wantSub:  "pothole",
		},
		{
			// Free text would make the eval set unscoreable: nothing could
			// decide whether the classifier's answer matched it.
			name:     "a subcategory outside the taxonomy is refused",
			body:     `{"label":"big hole innit","frame_type":"close"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "an empty label is refused",
			body:     `{"label":"","frame_type":"close"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "case and spacing are forgiven",
			body:     `{"label":"  Open Manhole ","frame_type":"close"}`,
			wantCode: http.StatusOK,
			wantSub:  "open manhole",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := &fakeLabels{}
			rec := httptest.NewRecorder()
			labelServer(t, labels, &fakeBlobs{}).Handler().
				ServeHTTP(rec, authed("POST", "/v1/label/r1", tt.body))

			if rec.Code != tt.wantCode {
				t.Fatalf("got %d, want %d: %s", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantCode != http.StatusOK {
				if len(labels.saved) != 0 {
					t.Error("a refused label must not be stored")
				}
				return
			}
			if len(labels.saved) != 1 {
				t.Fatalf("expected one saved label, got %d", len(labels.saved))
			}
			if got := labels.saved[0].Label; got != tt.wantSub {
				t.Errorf("stored %q, want %q", got, tt.wantSub)
			}
		})
	}
}

func TestLabellerSeesOnlyTheirOwnPhotographs(t *testing.T) {
	blobs := &fakeBlobs{put: map[string][]byte{"media/r1/a.jpg": []byte("jpegbytes")}}
	labels := &fakeLabels{
		keyBy:  map[string]string{"r1": "media/r1/a.jpg", "r2": "media/r2/a.jpg"},
		ownsBy: map[string]string{"r1": "account-1", "r2": "someone-else"},
	}
	h := labelServer(t, labels, blobs).Handler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authed("GET", "/v1/reports/r1/media", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("own photograph: got %d", rec.Code)
	}
	if rec.Body.String() != "jpegbytes" {
		t.Errorf("wrong bytes served")
	}

	// Someone else's capture is not found, not forbidden — a 403 confirms the
	// report exists, which is itself a disclosure.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, authed("GET", "/v1/reports/r2/media", ""))
	if rec.Code != http.StatusNotFound {
		t.Errorf("another account's photograph: got %d, want 404", rec.Code)
	}
}

func TestLabelQueueNeedsAuthentication(t *testing.T) {
	rec := httptest.NewRecorder()
	labelServer(t, &fakeLabels{}, &fakeBlobs{}).Handler().
		ServeHTTP(rec, httptest.NewRequest("GET", "/v1/label/queue", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", rec.Code)
	}
}

var _ = json.Marshal

// The capture path writes a label straight from the request, and until now
// nothing checked it. The field kit sent its own hardcoded vocabulary —
// "missing_manhole_cover", "garbage", "not_civic" — none of which the taxonomy
// contains, so a day of fieldwork would have produced labels the evaluation
// cannot score and every case would have counted as a miss.
func TestACaptureLabelMustBeInTheTaxonomy(t *testing.T) {
	tests := []struct {
		name  string
		label string
		kept  bool
	}{
		{name: "a taxonomy value is kept", label: "open manhole", kept: true},
		{name: "case and spacing are forgiven", label: "  Open Manhole ", kept: true},
		{name: "the field kit's underscored form is refused", label: "missing_manhole_cover"},
		{name: "a value that is not in the taxonomy is refused", label: "garbage"},
		{name: "the old not-civic marker is refused", label: "not_civic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reports := &fakeReports{}
			srv := newTestServer(t, reports, &fakeBlobs{})

			meta := `{"location":{"lat":19.23,"lon":72.84,"accuracy_m":6},` +
				`"captured_at":"2026-10-10T09:00:00Z",` +
				`"label":{"frame_type":"close","label":"` + tt.label + `"}}`
			req := captureRequest(t, meta, map[string][]byte{"close": []byte("jpegbytes")})
			req.Header.Set("Authorization", "Bearer test-token")
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, req)

			// The capture itself is never rejected over its label. Hard rule 7:
			// the photograph is the thing that cannot be retaken.
			if rec.Code != http.StatusAccepted {
				t.Fatalf("the capture was rejected: %d %s", rec.Code, rec.Body.String())
			}

			if tt.kept {
				if len(reports.labels) != 1 {
					t.Fatalf("a valid label was not stored: %+v", reports.labels)
				}
				if reports.labels[0].Label != strings.ToLower(strings.TrimSpace(tt.label)) {
					t.Errorf("stored %q", reports.labels[0].Label)
				}
				return
			}
			if len(reports.labels) != 0 {
				t.Errorf("an unscoreable label was stored: %q", reports.labels[0].Label)
			}
		})
	}
}
