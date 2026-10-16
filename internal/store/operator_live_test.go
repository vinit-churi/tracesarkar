package store_test

import (
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

// The point of this surface is that every number is real and every number is
// explained. A count nobody can act on, shown beside one somebody should,
// trains an operator to ignore both.
func TestLiveOperatorHealthAnswersEveryQuestionItAsks(t *testing.T) {
	ctx, _, db := liveDB(t)

	h, err := db.OperatorHealth(ctx, 5)
	if err != nil {
		t.Fatalf("OperatorHealth: %v", err)
	}
	if h.At.IsZero() {
		t.Error("the snapshot is not stamped")
	}
	if len(h.Queues) == 0 {
		t.Fatal("no queues reported")
	}
	for _, q := range h.Queues {
		if q.Name == "" {
			t.Error("a queue has no name")
		}
		if q.Clears == "" {
			t.Errorf("queue %q does not say what clears it", q.Name)
		}
		if q.Count < 0 {
			t.Errorf("queue %q counts %d", q.Name, q.Count)
		}
		t.Logf("%-42s %4d   clears: %s", q.Name, q.Count, q.Clears)
	}

	// Every status the table allows, including the ones at zero: a status
	// missing from the view reads as "none" and is indistinguishable from
	// "we stopped counting it".
	want := map[string]bool{"open": false, "claimed_resolved": false,
		"citizen_confirmed": false, "reopened": false}
	for _, c := range h.Issues {
		if _, ok := want[c.Status]; !ok {
			t.Errorf("unexpected status %q", c.Status)
			continue
		}
		want[c.Status] = true
		t.Logf("issues %-18s %4d", c.Status, c.Count)
	}
	for status, seen := range want {
		if !seen {
			t.Errorf("the view omits the %q status", status)
		}
	}

	claimed, confirmed := h.ClaimedVersusConfirmed()
	t.Logf("claimed_resolved %d, citizen_confirmed %d, captures %d, sources %d, runs %d",
		claimed, confirmed, h.Captures, len(h.Sources), len(h.Runs))

	// The queue counts cannot exceed the number of captures that could be in
	// them, which is the cheapest check that a query is counting the right
	// table.
	for _, q := range h.Queues {
		switch q.Name {
		case "captures not enriched", "captures not classified",
			"captures not attached to an issue":
			if q.Count > h.Captures {
				t.Errorf("%q counts %d of %d captures", q.Name, q.Count, h.Captures)
			}
		}
	}
}

// runLimit zero means "do not ask", because the runs list is the one part of
// this that costs a second query and is not always wanted.
func TestLiveOperatorHealthSkipsTheRunsListWhenNotAsked(t *testing.T) {
	ctx, _, db := liveDB(t)
	h, err := db.OperatorHealth(ctx, 0)
	if err != nil {
		t.Fatalf("OperatorHealth: %v", err)
	}
	if len(h.Runs) != 0 {
		t.Errorf("asked for no runs and got %d", len(h.Runs))
	}
	if len(h.Queues) == 0 {
		t.Error("the queues went missing with the runs")
	}
	var _ store.Health = h
}

// A queue that says a job will clear it has to count exactly what that job
// picks up.
//
// The first draft of this dashboard did not. It reported "captures not
// attached to an issue: 3" while the sweep's issues stage printed "nothing to
// attach", because the stage attaches a capture only when its latest
// classification was accepted or not_yet_covered and those three were
// needs_confirmation — waiting on the citizen, not on any job. Both numbers
// were right and the dashboard was misleading, which is the failure mode an
// operator console has.
func TestLiveAQueueCountsWhatItsStageActuallyPicksUp(t *testing.T) {
	ctx, _, db := liveDB(t)

	h, err := db.OperatorHealth(ctx, 0)
	if err != nil {
		t.Fatalf("OperatorHealth: %v", err)
	}
	var reported int
	var found bool
	for _, q := range h.Queues {
		if q.Name == "captures not attached to an issue" {
			reported, found = q.Count, true
		}
	}
	if !found {
		t.Fatal("the attachment queue is missing")
	}

	// The stage's own query, which is the authority.
	pending, err := db.UnattachedReports(ctx, 500)
	if err != nil {
		t.Fatalf("UnattachedReports: %v", err)
	}
	if reported != len(pending) {
		t.Errorf("the dashboard says %d captures await the issues stage; the stage "+
			"itself finds %d", reported, len(pending))
	}

	// Same for classification.
	for _, q := range h.Queues {
		if q.Name != "captures not classified" {
			continue
		}
		unclassified, err := db.UnclassifiedReports(ctx, 500)
		if err != nil {
			t.Fatalf("UnclassifiedReports: %v", err)
		}
		if q.Count != len(unclassified) {
			t.Errorf("the dashboard says %d captures await classification; the stage "+
				"itself finds %d", q.Count, len(unclassified))
		}
	}
}

// A sparse series lies by omission. A day on which collection did not run at
// all is the most important day in the window, and leaving it out draws it as
// if it never existed — so every day in the window is present, in order, with
// zeros where nothing happened.
func TestLiveTheTrendHasEveryDayInTheWindow(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, days := range []int{1, 7, 30} {
		tr, err := db.OperatorTrend(ctx, days)
		if err != nil {
			t.Fatalf("OperatorTrend(%d): %v", days, err)
		}
		if tr.Days != days {
			t.Errorf("asked for %d days and got %d", days, tr.Days)
		}
		for name, n := range map[string]int{
			"fetches": len(tr.Fetches), "runs": len(tr.Runs), "captures": len(tr.Captures),
		} {
			if n != days {
				t.Errorf("%d-day window: %s has %d buckets, want %d", days, name, n, days)
			}
		}
		// Same days, in the same order, across all three — a chart drawn from
		// three series with different axes is three charts.
		for i := range tr.Fetches {
			if i >= len(tr.Runs) || i >= len(tr.Captures) {
				break
			}
			if tr.Fetches[i].Day != tr.Runs[i].Day || tr.Fetches[i].Day != tr.Captures[i].Day {
				t.Errorf("bucket %d: fetches=%s runs=%s captures=%s",
					i, tr.Fetches[i].Day, tr.Runs[i].Day, tr.Captures[i].Day)
			}
		}
		// Oldest first, strictly increasing.
		for i := 1; i < len(tr.Fetches); i++ {
			if tr.Fetches[i].Day <= tr.Fetches[i-1].Day {
				t.Errorf("days are not oldest-first: %s then %s",
					tr.Fetches[i-1].Day, tr.Fetches[i].Day)
			}
		}
	}

	// A silly window falls back rather than scanning a decade.
	tr, err := db.OperatorTrend(ctx, 10000)
	if err != nil {
		t.Fatalf("OperatorTrend(10000): %v", err)
	}
	if tr.Days != 30 || len(tr.Fetches) != 30 {
		t.Errorf("an out-of-range window gave %d days and %d buckets",
			tr.Days, len(tr.Fetches))
	}
}

// Failures are counted, never inferred from a gap, and the three run outcomes
// stay separate: a process that died is not a job that returned an error.
func TestLiveTheTrendCountsWhatItSaysItCounts(t *testing.T) {
	ctx, pool, db := liveDB(t)

	tr, err := db.OperatorTrend(ctx, 30)
	if err != nil {
		t.Fatalf("OperatorTrend: %v", err)
	}

	var attempts, failures, runs int
	for _, f := range tr.Fetches {
		if f.Failures > f.Attempts {
			t.Errorf("%s: %d failures of %d attempts", f.Day, f.Failures, f.Attempts)
		}
		attempts += f.Attempts
		failures += f.Failures
	}
	for _, r := range tr.Runs {
		runs += r.OK + r.Failed + r.NeverReturned
	}

	// The totals have to match the tables, which is the cheapest check that
	// the day buckets are not dropping or double-counting rows.
	var wantFetches, wantRuns int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM fetch_log
		 WHERE (requested_at AT TIME ZONE 'UTC')::date
		       > (now() AT TIME ZONE 'UTC')::date - 30`).Scan(&wantFetches); err != nil {
		t.Fatalf("count fetches: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM collector_runs
		 WHERE (started_at AT TIME ZONE 'UTC')::date
		       > (now() AT TIME ZONE 'UTC')::date - 30`).Scan(&wantRuns); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if attempts != wantFetches {
		t.Errorf("the trend counts %d fetches in 30 days; the table has %d",
			attempts, wantFetches)
	}
	if runs != wantRuns {
		t.Errorf("the trend counts %d runs in 30 days; the table has %d", runs, wantRuns)
	}
	t.Logf("30 days: %d fetches (%d failed), %d runs", attempts, failures, runs)
	for _, r := range tr.Runs {
		if r.Failed > 0 || r.NeverReturned > 0 {
			t.Logf("  %s  ok=%d failed=%d never_returned=%d",
				r.Day, r.OK, r.Failed, r.NeverReturned)
		}
	}
}
