package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

// The labelling queue is what turns a day of walking into an eval set. It has
// to show each capture exactly once: a photograph labelled twice is a
// photograph someone wasted a second judgement on, and one never shown is a
// hole in the eval set nobody can see.
func TestLiveUnlabelledReportsDrainAsTheyAreLabelled(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM report_labels WHERE report_id IN
		    (SELECT id FROM reports WHERE account_id = $1)`, account)
		_, _ = pool.Exec(c, `DELETE FROM report_media WHERE report_id IN
		    (SELECT id FROM reports WHERE account_id = $1)`, account)
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	var ids []string
	for i := range 2 {
		captured := time.Now().UTC().Add(time.Duration(i) * time.Second)
		id, _, err := db.SaveReport(ctx, store.NewReport{
			AccountID: account, Lat: 19.2094, Lon: 72.8348, AccuracyM: 6,
			CapturedAt:  captured,
			Idempotency: fmt.Sprintf("label-%d-%s", i, captured.Format("150405.000000")),
		})
		if err != nil {
			t.Fatalf("SaveReport: %v", err)
		}
		if err := db.AddReportMedia(ctx, store.NewMedia{
			ReportID: id, ArchiveKey: "media/" + id + "/a.jpg",
			SHA256:      fmt.Sprintf("%064x", captured.UnixNano()+int64(i)),
			ContentType: "image/jpeg", Bytes: 1234, Role: "close",
			CapturedAt: &captured,
		}); err != nil {
			t.Fatalf("AddReportMedia: %v", err)
		}
		ids = append(ids, id)
	}

	mine := func(t *testing.T) []store.PendingLabel {
		t.Helper()
		all, err := db.UnlabelledReports(ctx, account, 500)
		if err != nil {
			t.Fatalf("UnlabelledReports: %v", err)
		}
		var out []store.PendingLabel
		for _, p := range all {
			for _, id := range ids {
				if p.ReportID == id {
					out = append(out, p)
				}
			}
		}
		return out
	}

	queued := mine(t)
	if len(queued) != 2 {
		t.Fatalf("both captures must be queued for labelling; got %d", len(queued))
	}
	// Only the requester's own captures. The media endpoint already refuses to
	// serve anyone else's photograph, so a queue that listed them would show a
	// column of broken images — and listing another account's report ids,
	// wards and timestamps is a disclosure in its own right.
	if len(queued) != len(mine(t)) {
		t.Fatal("the queue is not stable")
	}
	other, err := db.UnlabelledReports(ctx, "00000000-0000-0000-0000-000000000000", 500)
	if err != nil {
		t.Fatalf("UnlabelledReports for another account: %v", err)
	}
	for _, p := range other {
		for _, id := range ids {
			if p.ReportID == id {
				t.Errorf("capture %s was queued for an account that does not own it", id)
			}
		}
	}
	// A reporter id that is not a uuid at all — the static field-kit token —
	// must come back empty rather than erroring.
	if _, err := db.UnlabelledReports(ctx, "field-kit", 10); err != nil {
		t.Errorf("a non-uuid reporter must not error: %v", err)
	}
	// The archive key has to come with it, or the page cannot show the
	// photograph and the labeller is guessing.
	for _, p := range queued {
		if p.ArchiveKey == "" {
			t.Errorf("queued capture %s carries no media key", p.ReportID)
		}
	}

	if err := db.SaveReportLabel(ctx, store.ReportLabel{
		ReportID: ids[0], FrameType: "close", Label: "pothole",
		LabelledBy: account,
	}); err != nil {
		t.Fatalf("SaveReportLabel: %v", err)
	}

	left := mine(t)
	if len(left) != 1 {
		t.Fatalf("a labelled capture must leave the queue; %d still queued", len(left))
	}
	if left[0].ReportID != ids[1] {
		t.Errorf("the wrong capture left the queue")
	}

	// Relabelling must correct, not duplicate: a second judgement on the same
	// photograph replaces the first rather than giving the eval set two
	// conflicting answers for one image.
	if err := db.SaveReportLabel(ctx, store.ReportLabel{
		ReportID: ids[0], FrameType: "close", Label: "crack",
		LabelledBy: account,
	}); err != nil {
		t.Fatalf("relabel: %v", err)
	}
	var rows int
	var label string
	if err := pool.QueryRow(ctx,
		`SELECT count(*), max(label) FROM report_labels WHERE report_id = $1`,
		ids[0]).Scan(&rows, &label); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Errorf("relabelling duplicated: %d rows", rows)
	}
	if label != "crack" {
		t.Errorf("relabelling did not take: %q", label)
	}
}
