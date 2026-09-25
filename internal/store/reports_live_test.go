package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

func testAccount(t *testing.T, ctx context.Context, db *store.DB) string {
	t.Helper()
	id, err := db.EnsureAccount(ctx, "_test_"+time.Now().Format("150405.000000"))
	if err != nil {
		t.Fatalf("EnsureAccount: %v", err)
	}
	return id
}

func TestLiveSaveReportPersistsBeforeEnrichment(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	captured := time.Now().UTC().Add(-2 * time.Minute)
	in := store.NewReport{
		AccountID:   account,
		Lat:         19.2094,
		Lon:         72.8348,
		AccuracyM:   6.2,
		CapturedAt:  captured,
		Idempotency: "key-" + captured.Format("150405.000000"),
	}

	id, created, err := db.SaveReport(ctx, in)
	if err != nil {
		t.Fatalf("SaveReport: %v", err)
	}
	if !created {
		t.Fatal("the first save must create a report")
	}

	got, err := db.GetReport(ctx, id)
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if got.Status != "pending" {
		t.Errorf("a new report is pending until enrichment runs: got %q", got.Status)
	}
	// Coordinates must survive the round trip through PostGIS.
	if diff := got.Lat - 19.2094; diff > 0.00001 || diff < -0.00001 {
		t.Errorf("latitude drifted: got %v", got.Lat)
	}
	if diff := got.Lon - 72.8348; diff > 0.00001 || diff < -0.00001 {
		t.Errorf("longitude drifted: got %v", got.Lon)
	}
	if got.AccuracyM != 6.2 {
		t.Errorf("accuracy lost: got %v", got.AccuracyM)
	}
	// The gap between the device clock and ours is a cheap fraud signal.
	if got.ClockSkewSeconds < 100 {
		t.Errorf("clock skew should be recorded: got %d", got.ClockSkewSeconds)
	}
}

func TestLiveSaveReportIsIdempotent(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	in := store.NewReport{
		AccountID: account, Lat: 19.21, Lon: 72.83, AccuracyM: 8,
		CapturedAt: time.Now().UTC(), Idempotency: "the-same-key",
	}

	first, created, err := db.SaveReport(ctx, in)
	if err != nil || !created {
		t.Fatalf("first save: id=%s created=%v err=%v", first, created, err)
	}

	// A retried upload — the phone lost the response and sent it again.
	second, created, err := db.SaveReport(ctx, in)
	if err != nil {
		t.Fatalf("second save: %v", err)
	}
	if created {
		t.Error("a retry must not create a second report")
	}
	if second != first {
		t.Errorf("a retry must return the original report: %s vs %s", second, first)
	}
}

// A capture that is retried after a dropped connection re-uploads the same
// photograph. The report is idempotent; the media row must be too, or one
// capture accumulates a media row per retry.
func TestLiveReuploadingTheSamePhotographDoesNotDuplicateTheMediaRow(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	id, _, err := db.SaveReport(ctx, store.NewReport{
		AccountID: account, Lat: 19.21, Lon: 72.83, AccuracyM: 5,
		CapturedAt: time.Now().UTC(), Idempotency: "retry-media-test",
	})
	if err != nil {
		t.Fatalf("SaveReport: %v", err)
	}

	media := store.NewMedia{
		ReportID: id, Role: "close", ArchiveKey: "media/retry/close.jpg",
		ContentType: "image/jpeg", Bytes: 631,
		SHA256: "8b06656a27a75ec1d46ba8e4fac01f74bce056fcf40344d187e8e1de49973ebc",
	}
	for i := range 2 {
		if err := db.AddReportMedia(ctx, media); err != nil {
			t.Fatalf("AddReportMedia attempt %d: %v", i+1, err)
		}
	}

	got, err := db.GetReport(ctx, id)
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if len(got.Media) != 1 {
		t.Fatalf("one photograph must produce one media row, got %d", len(got.Media))
	}
}

func TestLiveAddMediaAndLabel(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	id, _, err := db.SaveReport(ctx, store.NewReport{
		AccountID: account, Lat: 19.21, Lon: 72.83, AccuracyM: 5,
		CapturedAt: time.Now().UTC(), Idempotency: "media-test",
	})
	if err != nil {
		t.Fatalf("SaveReport: %v", err)
	}

	if err := db.AddReportMedia(ctx, store.NewMedia{
		ReportID: id, Role: "close", ArchiveKey: "media/x/close.jpg",
		ContentType: "image/jpeg", Bytes: 1234,
		SHA256: "aa11bb22cc33dd44ee55ff6677889900aa11bb22cc33dd44ee55ff6677889900",
	}); err != nil {
		t.Fatalf("AddReportMedia: %v", err)
	}

	if err := db.SaveReportLabel(ctx, store.ReportLabel{
		ReportID: id, FrameType: "close", Label: "pothole",
		Conditions: []string{"day", "rain"}, WardGroundTruth: "R/S", LabelledBy: account,
	}); err != nil {
		t.Fatalf("SaveReportLabel: %v", err)
	}

	got, err := db.GetReport(ctx, id)
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if len(got.Media) != 1 || got.Media[0].Role != "close" {
		t.Fatalf("media not attached: %+v", got.Media)
	}
	if got.Label == nil || got.Label.Label != "pothole" || got.Label.WardGroundTruth != "R/S" {
		t.Fatalf("label not attached: %+v", got.Label)
	}

	// The eval set is exported by label, so counting by label must work.
	counts, err := db.LabelCounts(ctx)
	if err != nil {
		t.Fatalf("LabelCounts: %v", err)
	}
	if counts["pothole"] < 1 {
		t.Errorf("expected at least one pothole label, got %v", counts)
	}
}

func TestLiveSaveReportRejectsImpossibleCoordinates(t *testing.T) {
	ctx, _, db := liveDB(t)
	_, _, err := db.SaveReport(ctx, store.NewReport{
		AccountID: "", Lat: 200, Lon: 500, CapturedAt: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("coordinates outside the world must be refused")
	}
}
