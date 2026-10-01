package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/classify"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// A capture whose photograph can never be classified — corrupt bytes, a format
// the provider rejects — must stop being retried.
//
// Two things go wrong otherwise, and both bite hardest on a day of fieldwork.
// The queue is ordered oldest-first with a limit, so a permanently-failing
// capture sits at its head and crowds out everything taken after it. And each
// sweep pays for the same doomed vision call again, every ten minutes, forever.
//
// Hard rule 7 requires every stage to be retryable. It does not require them to
// be retried without end.
func TestLiveAPermanentlyFailingCaptureStopsBeingRetried(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM classifications WHERE report_id IN
		    (SELECT id FROM reports WHERE account_id = $1)`, account)
		_, _ = pool.Exec(c, `DELETE FROM report_media WHERE report_id IN
		    (SELECT id FROM reports WHERE account_id = $1)`, account)
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	captured := time.Now().UTC()
	id, _, err := db.SaveReport(ctx, store.NewReport{
		AccountID:   account,
		Lat:         19.2094,
		Lon:         72.8348,
		AccuracyM:   6,
		CapturedAt:  captured,
		Idempotency: "retrycap-" + captured.Format("150405.000000"),
	})
	if err != nil {
		t.Fatalf("SaveReport: %v", err)
	}
	if err := db.AddReportMedia(ctx, store.NewMedia{
		ReportID:   id,
		ArchiveKey: "media/" + id + "/broken.jpg",
		// Content-addressed keys are hex digests; a unique one per run keeps
		// the media insert from deduplicating against an earlier test.
		SHA256:      fmt.Sprintf("%064x", captured.UnixNano()),
		ContentType: "image/jpeg",
		Bytes:       20,
		Role:        "close",
		CapturedAt:  &captured,
	}); err != nil {
		t.Fatalf("AddReportMedia: %v", err)
	}

	contains := func(t *testing.T, limit int) bool {
		t.Helper()
		pending, err := db.Pending(ctx, limit)
		if err != nil {
			t.Fatalf("Pending: %v", err)
		}
		for _, p := range pending {
			if p.ReportID == id {
				return true
			}
		}
		return false
	}

	if !contains(t, 200) {
		t.Fatal("a capture with no classification attempt must be pending")
	}

	// Fail it up to the cap. Each one must leave it retryable.
	for i := 1; i < store.MaxClassifyAttempts; i++ {
		if err := db.Save(ctx, classify.Attempt{
			ReportID: id, Model: "test", Prompt: "v1", Err: "unsupported image",
		}); err != nil {
			t.Fatalf("save attempt %d: %v", i, err)
		}
		if !contains(t, 200) {
			t.Fatalf("after %d of %d failures the capture must still be retried",
				i, store.MaxClassifyAttempts)
		}
	}

	// The one that reaches the cap retires it.
	if err := db.Save(ctx, classify.Attempt{
		ReportID: id, Model: "test", Prompt: "v1", Err: "unsupported image",
	}); err != nil {
		t.Fatalf("save final attempt: %v", err)
	}
	if contains(t, 200) {
		t.Errorf("after %d failures the capture must stop being retried",
			store.MaxClassifyAttempts)
	}
}
