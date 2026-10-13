package store_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

func TestLiveRedactionTargetsAreThePhotographsNothingHasLookedAt(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM report_media WHERE report_id IN
			(SELECT id FROM reports WHERE account_id = $1)`, account)
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	stamp := time.Now().UTC()
	reportID, _, err := db.SaveReport(ctx, store.NewReport{
		AccountID: account, Lat: 19.2094, Lon: 72.8348, AccuracyM: 6,
		CapturedAt: stamp, Idempotency: "redact-" + stamp.Format("150405.000000000"),
	})
	if err != nil {
		t.Fatalf("SaveReport: %v", err)
	}

	digest := sha256.Sum256([]byte("redaction-test-" + reportID))
	if err := db.AddReportMedia(ctx, store.NewMedia{
		ReportID: reportID, Role: "close", ArchiveKey: "media/" + reportID + "/x",
		ContentType: "image/jpeg", Bytes: 1234, SHA256: hex.EncodeToString(digest[:]),
		CapturedAt: &stamp,
	}); err != nil {
		t.Fatalf("AddReportMedia: %v", err)
	}

	mine := func(t *testing.T) store.RedactionTarget {
		t.Helper()
		targets, err := db.MediaToRedact(ctx, 500)
		if err != nil {
			t.Fatalf("MediaToRedact: %v", err)
		}
		for _, tg := range targets {
			if tg.ReportID == reportID {
				return tg
			}
		}
		t.Fatal("the new photograph was not offered for redaction")
		return store.RedactionTarget{}
	}

	target := mine(t)
	if target.ArchiveKey == "" {
		t.Fatal("target carries no archive key; there would be nothing to fetch")
	}

	// Holding takes the photograph out of the queue — it is a decision, not a
	// deferral — and still records where the redacted copy went, because that
	// copy is what the classifier reads.
	if err := db.RecordRedaction(ctx, target.MediaID, store.RedactionResult{
		RanAt:     time.Now().UTC(),
		Tool:      "pigo-1.4.6/facefinder+edge-density-1",
		OutputKey: "derivative/" + reportID + "/held.jpg",
		Reason:    "people were seen in this photograph but no face was found to obscure",
	}); err != nil {
		t.Fatalf("RecordRedaction (held): %v", err)
	}

	targets, err := db.MediaToRedact(ctx, 500)
	if err != nil {
		t.Fatalf("MediaToRedact after holding: %v", err)
	}
	for _, tg := range targets {
		if tg.ReportID == reportID {
			t.Fatal("a photograph with a verdict was offered for redaction again")
		}
	}

	var verdict, reason string
	var outputKey *string
	if err := pool.QueryRow(ctx, `
		SELECT redaction_verdict, COALESCE(redaction_reason,''), redaction_output_key
		FROM report_media WHERE id = $1::uuid`, target.MediaID).
		Scan(&verdict, &reason, &outputKey); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if verdict != "held" {
		t.Fatalf("verdict = %q, want held", verdict)
	}
	if outputKey == nil {
		t.Fatal("a held photograph has no redacted copy; the classifier would have to read the original")
	}
	if reason == "" {
		t.Fatal("held with no reason recorded; nothing could explain the gap to a reporter")
	}
}

func TestLiveRecordRedactionRefusesAnyVerdictWithoutAnOutputObject(t *testing.T) {
	_, _, db := liveDB(t)
	for _, published := range []bool{true, false} {
		err := db.RecordRedaction(context.Background(), "00000000-0000-0000-0000-000000000000",
			store.RedactionResult{RanAt: time.Now(), Tool: "t", Published: published})
		if err == nil {
			t.Fatalf("a verdict with no output object was accepted (published=%v)", published)
		}
	}
}

func TestLiveRecordRedactionOnAVanishedRowSaysSo(t *testing.T) {
	ctx, _, db := liveDB(t)
	err := db.RecordRedaction(ctx, "00000000-0000-0000-0000-000000000000",
		store.RedactionResult{
			RanAt: time.Now(), Tool: "t", OutputKey: "k", Published: true,
		})
	if !errors.Is(err, store.ErrMediaGone) {
		t.Fatalf("err = %v, want ErrMediaGone — a deleted photograph must not look like success", err)
	}
}
