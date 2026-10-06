package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/classify"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// The evaluation set is every labelled capture, read back as the cases the
// harness scores. Until this existed a weekend of walking produced labelled
// photographs and still no number.
//
// Ground truth is one subcategory, and the category and hazard flag are derived
// from the taxonomy rather than stored again — three fields that can disagree
// eventually will, and a wrong hazard flag silently moves the recall figure the
// model is judged against.
func TestLiveEvalSetIsBuiltFromTheLabels(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		for _, q := range []string{
			`DELETE FROM report_labels WHERE report_id IN (SELECT id FROM reports WHERE account_id=$1)`,
			`DELETE FROM report_media WHERE report_id IN (SELECT id FROM reports WHERE account_id=$1)`,
			`DELETE FROM reports WHERE account_id=$1`,
			`DELETE FROM accounts WHERE id=$1`,
		} {
			_, _ = pool.Exec(c, q, account)
		}
	})

	// One ordinary capture and one hazard, so the derivation is exercised.
	want := []struct {
		label, category string
		hazard          bool
	}{
		{"pothole", "road_defect", false},
		{"open manhole", "road_defect", true},
	}
	ids := make([]string, 0, len(want))
	for i, w := range want {
		captured := time.Now().UTC().Add(time.Duration(i) * time.Second)
		id, _, err := db.SaveReport(ctx, store.NewReport{
			AccountID: account, Lat: 19.2340, Lon: 72.8444, AccuracyM: 6,
			CapturedAt:  captured,
			Idempotency: fmt.Sprintf("eval-%d-%s", i, captured.Format("150405.000000")),
		})
		if err != nil {
			t.Fatalf("SaveReport: %v", err)
		}
		if err := db.AddReportMedia(ctx, store.NewMedia{
			ReportID: id, ArchiveKey: "media/" + id + "/a.jpg",
			SHA256:      fmt.Sprintf("%064x", captured.UnixNano()+int64(i)),
			ContentType: "image/jpeg", Bytes: 10, Role: "close",
			CapturedAt: &captured,
		}); err != nil {
			t.Fatalf("AddReportMedia: %v", err)
		}
		if err := db.SaveReportLabel(ctx, store.ReportLabel{
			ReportID: id, FrameType: "close", Label: w.label,
			Conditions: []string{"day"}, Notes: "taken at noon",
			LabelledBy: account,
		}); err != nil {
			t.Fatalf("SaveReportLabel: %v", err)
		}
		ids = append(ids, id)
	}

	all, err := db.EvalSet(ctx, 500)
	if err != nil {
		t.Fatalf("EvalSet: %v", err)
	}
	got := map[string]store.EvalEntry{}
	for _, e := range all {
		for _, id := range ids {
			if e.ReportID == id {
				got[e.Label] = e
			}
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected both labelled captures in the eval set, got %d", len(got))
	}

	for _, w := range want {
		e, ok := got[w.label]
		if !ok {
			t.Fatalf("%q missing from the eval set", w.label)
		}
		if classify.CategoryOf(e.Label) != w.category {
			t.Errorf("%q derives category %q, want %q",
				w.label, classify.CategoryOf(e.Label), w.category)
		}
		if classify.IsHazard(e.Label) != w.hazard {
			t.Errorf("%q derives hazard %v, want %v",
				w.label, classify.IsHazard(e.Label), w.hazard)
		}
		// The archive key has to come with it, or the harness has no image.
		if e.ArchiveKey == "" {
			t.Errorf("%q carries no media key", w.label)
		}
		// The labeller's note is what usually explains a miss.
		if e.Notes != "taken at noon" {
			t.Errorf("%q lost its note: %q", w.label, e.Notes)
		}
		if len(e.Conditions) == 0 {
			t.Errorf("%q lost its conditions", w.label)
		}
	}
}
