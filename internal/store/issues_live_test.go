package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/issue"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// The whole of system 4 in one test: two people photograph the same pothole
// and the platform records one problem seen twice, not two problems.
func TestLiveASecondSightingJoinsTheFirst(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM issues WHERE id IN
		   (SELECT issue_id FROM issue_reports WHERE report_id IN
		     (SELECT id FROM reports WHERE account_id = $1))`, account)
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	report := func(i int) string {
		captured := time.Now().UTC().Add(time.Duration(i) * time.Second)
		id, _, err := db.SaveReport(ctx, store.NewReport{
			AccountID: account, Lat: 19.2340, Lon: 72.8444, AccuracyM: 6,
			CapturedAt:  captured,
			Idempotency: fmt.Sprintf("issue-%d-%s", i, captured.Format("150405.000000")),
		})
		if err != nil {
			t.Fatalf("SaveReport: %v", err)
		}
		return id
	}

	attach := func(reportID string, lat, lon float64) (string, issue.Decision) {
		near, err := db.IssuesNear(ctx, lat, lon, 100)
		if err != nil {
			t.Fatalf("IssuesNear: %v", err)
		}
		cands := make([]issue.Candidate, 0, len(near))
		for _, n := range near {
			cands = append(cands, issue.Candidate{
				IssueID: n.IssueID, Category: n.Category,
				Status: n.Status, DistanceM: n.DistanceM,
			})
		}
		d := issue.Decide(6, "road_defect", cands)
		id, err := db.AttachReport(ctx, store.Attachment{
			ReportID: reportID, Authority: "BMC", Ward: "R/C",
			Category: "road_defect", Subcategory: "pothole",
			Lat: lat, Lon: lon, SeenAt: time.Now().UTC(),
			Actor: account, Decision: d,
		})
		if err != nil {
			t.Fatalf("AttachReport: %v", err)
		}
		return id, d
	}

	first, d1 := attach(report(1), 19.2340, 72.8444)
	if d1.Action != issue.Create {
		t.Fatalf("the first sighting should create: %q", d1.Action)
	}

	// Four metres away — the same pothole, photographed from the other side.
	second, d2 := attach(report(2), 19.23403, 72.84442)
	if d2.Action != issue.Join {
		t.Fatalf("the second sighting should join: %q — %s", d2.Action, d2.Basis)
	}
	if second != first {
		t.Fatalf("a second issue was created: %s then %s", first, second)
	}

	var sightings int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM issue_reports WHERE issue_id = $1::uuid`, first).
		Scan(&sightings); err != nil {
		t.Fatal(err)
	}
	if sightings != 2 {
		t.Errorf("the issue has %d sightings, want 2", sightings)
	}

	// The history is the evidence, and it must verify as written.
	events, err := db.IssueTimeline(ctx, first)
	if err != nil {
		t.Fatalf("IssueTimeline: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("the history has %d events, want 2", len(events))
	}
	if events[0].Kind != "reported" || events[1].Kind != "corroborated" {
		t.Errorf("the history reads %q then %q", events[0].Kind, events[1].Kind)
	}
	if err := issue.Verify(events); err != nil {
		t.Errorf("the chain as stored does not verify: %v", err)
	}
	// And the chain must actually be chained, not a row of unrelated hashes.
	if events[1].PrevHash != events[0].Hash {
		t.Error("the second event does not point at the first")
	}
}

// A hundred metres away is a different pothole, and merging them would hide
// one of them.
func TestLiveADistantSightingIsItsOwnIssue(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM issues WHERE id IN
		   (SELECT issue_id FROM issue_reports WHERE report_id IN
		     (SELECT id FROM reports WHERE account_id = $1))`, account)
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	mk := func(i int, lat, lon float64) string {
		captured := time.Now().UTC().Add(time.Duration(i) * time.Second)
		id, _, err := db.SaveReport(ctx, store.NewReport{
			AccountID: account, Lat: lat, Lon: lon, AccuracyM: 6,
			CapturedAt:  captured,
			Idempotency: fmt.Sprintf("far-%d-%s", i, captured.Format("150405.000000")),
		})
		if err != nil {
			t.Fatal(err)
		}
		near, _ := db.IssuesNear(ctx, lat, lon, 100)
		cands := make([]issue.Candidate, 0, len(near))
		for _, n := range near {
			cands = append(cands, issue.Candidate{
				IssueID: n.IssueID, Category: n.Category,
				Status: n.Status, DistanceM: n.DistanceM,
			})
		}
		issueID, err := db.AttachReport(ctx, store.Attachment{
			ReportID: id, Authority: "BMC", Ward: "R/C", Category: "road_defect",
			Lat: lat, Lon: lon, SeenAt: time.Now().UTC(), Actor: account,
			Decision: issue.Decide(6, "road_defect", cands),
		})
		if err != nil {
			t.Fatal(err)
		}
		return issueID
	}

	a := mk(1, 19.2340, 72.8444)
	b := mk(2, 19.2350, 72.8444) // ~110 m north
	if a == b {
		t.Error("two potholes a hundred metres apart were merged into one issue")
	}
}
