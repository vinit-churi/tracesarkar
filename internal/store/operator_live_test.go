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
