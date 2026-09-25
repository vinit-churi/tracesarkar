package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeReports stands in for the database; the store's own behaviour is covered
// by its live tests.
type fakeReports struct {
	saved      []NewReport
	media      []NewMedia
	nextID     string
	created    bool
	saveErr    error
	mediaErr   error
	labelSaved int
}

func (f *fakeReports) SaveReport(_ context.Context, in NewReport) (string, bool, error) {
	if f.saveErr != nil {
		return "", false, f.saveErr
	}
	f.saved = append(f.saved, in)
	id := f.nextID
	if id == "" {
		id = "report-1"
	}
	return id, !f.created, nil
}

func (f *fakeReports) AddReportMedia(_ context.Context, in NewMedia) error {
	if f.mediaErr != nil {
		return f.mediaErr
	}
	f.media = append(f.media, in)
	return nil
}

func (f *fakeReports) SaveReportLabel(_ context.Context, _ ReportLabel) error {
	f.labelSaved++
	return nil
}

type fakeBlobs struct {
	put map[string][]byte
	err error
}

func (f *fakeBlobs) Put(_ context.Context, key string, body []byte, _ string) error {
	if f.err != nil {
		return f.err
	}
	if f.put == nil {
		f.put = map[string][]byte{}
	}
	f.put[key] = body
	return nil
}

func newTestServer(t *testing.T, reports *fakeReports, blobs *fakeBlobs) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: reports,
		Media:   blobs,
		Token:   "test-token",
		Account: "account-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

// captureRequest builds the multipart body a phone sends.
func captureRequest(t *testing.T, meta string, images map[string][]byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if meta != "" {
		if err := w.WriteField("meta", meta); err != nil {
			t.Fatal(err)
		}
	}
	for role, content := range images {
		part, err := w.CreateFormFile("media", role+".jpg")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatal(err)
		}
		if err := w.WriteField("role", role); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/reports", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-token")
	return req
}

const validMeta = `{"location":{"lat":19.2094,"lon":72.8348,"accuracy_m":6.2},
                    "captured_at":"2026-09-25T08:14:22+05:30"}`

func TestPostReportStoresTheCaptureAndReturns202(t *testing.T) {
	reports, blobs := &fakeReports{}, &fakeBlobs{}
	srv := newTestServer(t, reports, blobs)

	req := captureRequest(t, validMeta, map[string][]byte{"close": []byte("jpeg-bytes")})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got %d, want 202: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["report_id"] != "report-1" {
		t.Errorf("report_id: got %v", resp["report_id"])
	}
	if resp["status"] != "pending" {
		t.Errorf("status: got %v — enrichment has not run yet", resp["status"])
	}

	if len(reports.saved) != 1 {
		t.Fatalf("expected one saved report, got %d", len(reports.saved))
	}
	saved := reports.saved[0]
	if saved.Lat != 19.2094 || saved.Lon != 72.8348 {
		t.Errorf("coordinates: %+v", saved)
	}
	if saved.AccuracyM != 6.2 {
		t.Errorf("accuracy must survive: got %v", saved.AccuracyM)
	}
}

func TestPostReportArchivesTheImageBeforeReplying(t *testing.T) {
	reports, blobs := &fakeReports{}, &fakeBlobs{}
	srv := newTestServer(t, reports, blobs)

	req := captureRequest(t, validMeta, map[string][]byte{"close": []byte("jpeg-bytes")})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	if len(blobs.put) != 1 {
		t.Fatalf("the photograph must be stored, got %d objects", len(blobs.put))
	}
	for key, body := range blobs.put {
		if !strings.HasPrefix(key, "media/") {
			t.Errorf("media belongs under media/: %s", key)
		}
		if string(body) != "jpeg-bytes" {
			t.Errorf("stored bytes differ from what was uploaded")
		}
	}
	if len(reports.media) != 1 || reports.media[0].Role != "close" {
		t.Errorf("media row: %+v", reports.media)
	}
}

func TestPostReportRefusesWhenTheImageCannotBeStored(t *testing.T) {
	reports := &fakeReports{}
	blobs := &fakeBlobs{err: fmt.Errorf("bucket unreachable")}
	srv := newTestServer(t, reports, blobs)

	req := captureRequest(t, validMeta, map[string][]byte{"close": []byte("jpeg")})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	// Losing a capture is forbidden: if the photograph cannot be stored, the
	// client must be told so it can retry, not told everything is fine.
	if rec.Code < 500 {
		t.Fatalf("got %d, want a server error the client will retry", rec.Code)
	}
}

func TestPostReportRequiresLocation(t *testing.T) {
	srv := newTestServer(t, &fakeReports{}, &fakeBlobs{})

	req := captureRequest(t, `{"captured_at":"2026-09-25T08:14:22+05:30"}`,
		map[string][]byte{"close": []byte("jpeg")})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "location") {
		t.Errorf("the error should say what was missing: %s", rec.Body.String())
	}
}

func TestPostReportRequiresAtLeastOneImage(t *testing.T) {
	srv := newTestServer(t, &fakeReports{}, &fakeBlobs{})

	req := captureRequest(t, validMeta, nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func TestPostReportRejectsAnUnauthenticatedCaller(t *testing.T) {
	srv := newTestServer(t, &fakeReports{}, &fakeBlobs{})

	req := captureRequest(t, validMeta, map[string][]byte{"close": []byte("jpeg")})
	req.Header.Del("Authorization")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestPostReportPassesTheIdempotencyKeyThrough(t *testing.T) {
	reports := &fakeReports{}
	srv := newTestServer(t, reports, &fakeBlobs{})

	req := captureRequest(t, validMeta, map[string][]byte{"close": []byte("jpeg")})
	req.Header.Set("Idempotency-Key", "abc123")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	if reports.saved[0].Idempotency != "abc123" {
		t.Errorf("idempotency key lost: %+v", reports.saved[0])
	}
}

func TestPostReportStoresALabelWhenTheFieldKitSendsOne(t *testing.T) {
	reports := &fakeReports{}
	srv := newTestServer(t, reports, &fakeBlobs{})

	meta := `{"location":{"lat":19.2,"lon":72.8,"accuracy_m":5},
	          "captured_at":"2026-09-25T08:14:22+05:30",
	          "label":{"frame_type":"close","label":"pothole",
	                   "conditions":["night","rain"],"ward_ground_truth":"R/S"}}`
	req := captureRequest(t, meta, map[string][]byte{"close": []byte("jpeg")})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	if reports.labelSaved != 1 {
		t.Errorf("the field kit's label should be stored: %d", reports.labelSaved)
	}
}

func TestHealthEndpointNeedsNoToken(t *testing.T) {
	srv := newTestServer(t, &fakeReports{}, &fakeBlobs{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
}

func TestPostReportRejectsAnOversizedUpload(t *testing.T) {
	reports := &fakeReports{}
	srv, err := New(Options{
		Reports: reports, Media: &fakeBlobs{}, Token: "test-token",
		Account: "account-1", MaxUploadBytes: 1024,
	})
	if err != nil {
		t.Fatal(err)
	}

	big := bytes.Repeat([]byte("x"), 4096)
	req := captureRequest(t, validMeta, map[string][]byte{"close": big})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge && rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want the upload refused", rec.Code)
	}
	if len(reports.saved) != 0 {
		t.Error("nothing should be stored for a refused upload")
	}
}

func TestCapturedAtMustBeParseable(t *testing.T) {
	srv := newTestServer(t, &fakeReports{}, &fakeBlobs{})

	req := captureRequest(t, `{"location":{"lat":19.2,"lon":72.8},"captured_at":"yesterday"}`,
		map[string][]byte{"close": []byte("jpeg")})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
}
