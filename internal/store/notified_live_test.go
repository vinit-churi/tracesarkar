package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/issue"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Lodging the complaint is what starts the clock. Until it happens nothing is
// running, however long ago the photograph was taken.
func TestLiveLodgingStartsTheClockOnceAndOnlyOnce(t *testing.T) {
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

	captured := time.Now().UTC()
	reportID, _, err := db.SaveReport(ctx, store.NewReport{
		AccountID: account, Lat: 19.2341, Lon: 72.8445, AccuracyM: 6,
		CapturedAt:  captured,
		Idempotency: "notify-" + captured.Format("150405.000000"),
	})
	if err != nil {
		t.Fatal(err)
	}
	issueID, err := db.AttachReport(ctx, store.Attachment{
		ReportID: reportID, Authority: "BMC", Ward: "R/C",
		Category: "road_defect", Lat: 19.2341, Lon: 72.8445,
		SeenAt: captured, Actor: account,
		Decision: issue.Decision{Action: issue.Create, Basis: "first sighting"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Before: nothing running, whatever the deadline is.
	v, _, err := db.IssueForReport(ctx, reportID)
	if err != nil {
		t.Fatal(err)
	}
	if v.NotifiedAt != nil {
		t.Fatal("a clock was running before anybody was told")
	}
	if got := issue.NewClock(v.NotifiedAt, 48).At(time.Now()).State; got != issue.NotStarted {
		t.Errorf("clock state = %q before lodging, want not_started", got)
	}

	lodged := time.Now().UTC()
	if err := db.RecordNotified(ctx, issueID, "MARG-2026-1012-0001", account, lodged); err != nil {
		t.Fatalf("RecordNotified: %v", err)
	}

	// A second lodging must not move the start, or the deadline could be
	// pushed back indefinitely by re-submitting.
	err = db.RecordNotified(ctx, issueID, "MARG-LATER", account, lodged.Add(20*time.Hour))
	if !errors.Is(err, store.ErrAlreadyNotified) {
		t.Errorf("a second lodging was accepted: %v", err)
	}

	v, _, err = db.IssueForReport(ctx, reportID)
	if err != nil {
		t.Fatal(err)
	}
	if v.NotifiedAt == nil {
		t.Fatal("the clock did not start")
	}
	if v.Reference != "MARG-2026-1012-0001" {
		t.Errorf("reference = %q, want the first one", v.Reference)
	}
	reading := issue.NewClock(v.NotifiedAt, 48).At(lodged.Add(10 * time.Hour))
	if reading.State != issue.Running {
		t.Errorf("clock state = %q after lodging, want running", reading.State)
	}

	// The lodging is in the history, and the history still verifies.
	events, err := db.IssueTimeline(ctx, issueID)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, e := range events {
		if e.Kind == "notified" {
			found = true
			if e.Detail == "" {
				t.Error("the lodging carries no reference in the history")
			}
		}
	}
	if !found {
		t.Error("the lodging is not in the issue's history")
	}
	if err := issue.Verify(events); err != nil {
		t.Errorf("the chain does not verify after lodging: %v", err)
	}
}

// Only somebody who photographed it may say it has been lodged: the reference
// dates a legal clock.
func TestLiveOnlyAReporterMaySayItWasLodged(t *testing.T) {
	ctx, pool, db := liveDB(t)
	mine := testAccount(t, ctx, db)
	theirs := testAccount(t, ctx, db)

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		for _, a := range []string{mine, theirs} {
			_, _ = pool.Exec(c, `DELETE FROM issues WHERE id IN
			   (SELECT issue_id FROM issue_reports WHERE report_id IN
			     (SELECT id FROM reports WHERE account_id = $1))`, a)
			_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, a)
			_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, a)
		}
	})

	captured := time.Now().UTC()
	reportID, _, err := db.SaveReport(ctx, store.NewReport{
		AccountID: mine, Lat: 19.2342, Lon: 72.8446, AccuracyM: 6,
		CapturedAt:  captured,
		Idempotency: fmt.Sprintf("auth-%s", captured.Format("150405.000000")),
	})
	if err != nil {
		t.Fatal(err)
	}
	issueID, err := db.AttachReport(ctx, store.Attachment{
		ReportID: reportID, Category: "road_defect",
		Lat: 19.2342, Lon: 72.8446, SeenAt: captured, Actor: mine,
		Decision: issue.Decision{Action: issue.Create, Basis: "first"},
	})
	if err != nil {
		t.Fatal(err)
	}

	ok, err := db.ReporterOf(ctx, issueID, mine)
	if err != nil || !ok {
		t.Errorf("the person who photographed it is not recognised: %v %v", ok, err)
	}
	ok, err = db.ReporterOf(ctx, issueID, theirs)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("somebody who never photographed it may date the clock")
	}
}
