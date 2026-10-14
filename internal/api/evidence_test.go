package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/evidence"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

type fakeEvidence struct {
	record evidence.Record
	photos []store.EvidencePhoto
	found  bool
	gotAcc string
}

func (f *fakeEvidence) EvidenceFor(_ context.Context, _, accountID string) (evidence.Record, []store.EvidencePhoto, bool, error) {
	f.gotAcc = accountID
	return f.record, f.photos, f.found, nil
}

func anEvidenceRecord() evidence.Record {
	lodged := time.Date(2026, 10, 12, 11, 0, 0, 0, time.UTC)
	return evidence.Record{
		IssueID:  "3f1a0f9c-0000-4000-8000-000000000001",
		Category: "road_defect", Subcategory: "pothole", Status: "open",
		Authority: "BMC", Ward: "R/C", Sightings: 2,
		Lat: exactLat, Lon: exactLon, AccuracyM: 6.2,
		FirstSeen: time.Date(2026, 10, 11, 9, 14, 0, 0, time.UTC),
		LastSeen:  time.Date(2026, 10, 12, 8, 2, 0, 0, time.UTC),
		LodgedOn:  &lodged, Reference: reference,
	}
}

func evidenceServer(t *testing.T, e Evidence) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Evidence: e, Blobs: &fakeStore{},
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func getPack(t *testing.T, e Evidence, auth bool) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("GET",
		"/v1/issues/3f1a0f9c-0000-4000-8000-000000000001/evidence.pdf", nil)
	if auth {
		req.Header.Set("Authorization", "Bearer test-token")
	}
	rec := httptest.NewRecorder()
	evidenceServer(t, e).Handler().ServeHTTP(rec, req)
	return rec
}

func TestThePackIsServedAsAPDF(t *testing.T) {
	rec := getPack(t, &fakeEvidence{record: anEvidenceRecord(), found: true}, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type is %q", ct)
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF-")) {
		t.Error("the body is not a PDF")
	}
	// A filename a person can find again in a downloads folder.
	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "attachment") || !strings.Contains(cd, ".pdf") {
		t.Errorf("Content-Disposition is %q", cd)
	}
}

// Unlike every other share surface this one is not public: it carries the
// exact position and the complaint reference.
func TestThePackIsNotServedWithoutASignIn(t *testing.T) {
	rec := getPack(t, &fakeEvidence{record: anEvidenceRecord(), found: true}, false)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", rec.Code)
	}
}

// Scoped by sighting, and a stranger is told the issue does not exist rather
// than that they may not have it: which issues exist is itself information.
func TestSomebodyWithNoSightingIsToldTheIssueDoesNotExist(t *testing.T) {
	rec := getPack(t, &fakeEvidence{found: false}, true)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404", rec.Code)
	}
	if rec.Code == http.StatusForbidden {
		t.Error("403 confirms the issue exists")
	}
}

// The account is taken from the request, never from anything the caller sends.
func TestThePackIsScopedToTheAccountBehindTheRequest(t *testing.T) {
	f := &fakeEvidence{record: anEvidenceRecord(), found: true}
	getPack(t, f, true)
	if f.gotAcc == "" {
		t.Error("the handler asked for a pack with no account")
	}
}

// The pack is built from the redacted copies only. A photograph the gate held
// has no entry, and the handler must not reach for the original.
func TestOnlyRedactedCopiesReachThePack(t *testing.T) {
	f := &fakeEvidence{
		record: anEvidenceRecord(), found: true,
		photos: []store.EvidencePhoto{
			{Key: "derivative/r1/m1.jpg", Tool: "pigo-1.4.6/facefinder+edge-density-2",
				Regions: 1, Published: true},
		},
	}
	blobs := &fakeStore{}
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Evidence: f, Blobs: blobs,
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET",
		"/v1/issues/3f1a0f9c-0000-4000-8000-000000000001/evidence.pdf", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.HasPrefix(blobs.got, "derivative/") {
		t.Errorf("the pack read %q, which is not a redacted copy", blobs.got)
	}
}

// A photograph that will not decode must not fail the whole pack: the record
// is worth having without the picture.
func TestAnUnreadablePhotographCostsThePictureNotThePack(t *testing.T) {
	f := &fakeEvidence{
		record: anEvidenceRecord(), found: true,
		photos: []store.EvidencePhoto{{Key: "derivative/r1/m1.jpg", Published: true}},
	}
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Evidence: f, Blobs: &fakeStore{corrupt: true},
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET",
		"/v1/issues/3f1a0f9c-0000-4000-8000-000000000001/evidence.pdf", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF-")) {
		t.Error("the pack was lost along with the photograph")
	}
}

func TestThePackSaysSoWhenItIsNotConfigured(t *testing.T) {
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET",
		"/v1/issues/3f1a0f9c-0000-4000-8000-000000000001/evidence.pdf", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("status %d, want 501", rec.Code)
	}
}
