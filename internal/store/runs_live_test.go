package store_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

// A night that fails leaves no serial console behind, so the outcome of every
// run has to be durable somewhere we can query later.
func TestLiveRunLogRecordsSuccessAndFailure(t *testing.T) {
	ctx, pool, db := liveDB(t)

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM collector_runs WHERE host = '_test_host'`)
	})

	id, err := db.StartRun(ctx, store.RunStart{
		Command: "run", Host: "_test_host", Tier: "personal", Version: "test",
	})
	if err != nil {
		t.Fatalf("StartRun: %v", err)
	}
	if id == "" {
		t.Fatal("expected a run id")
	}

	// An unfinished run is visible, so a run that never returns is not invisible.
	runs, err := db.RecentRuns(ctx, 20)
	if err != nil {
		t.Fatalf("RecentRuns: %v", err)
	}
	var found *store.RunRow
	for i := range runs {
		if runs[i].ID == id {
			found = &runs[i]
		}
	}
	if found == nil {
		t.Fatal("a started run must be visible before it finishes")
	}
	if found.FinishedAt != nil {
		t.Error("it has not finished yet")
	}

	if err := db.FinishRun(ctx, id, store.RunFinish{
		OK: false, Error: "bmc_roads_api/publicdashboard: i/o timeout",
		Detail: map[string]any{"changed": 0, "failed": 3},
	}); err != nil {
		t.Fatalf("FinishRun: %v", err)
	}

	runs, err = db.RecentRuns(ctx, 20)
	if err != nil {
		t.Fatalf("RecentRuns after finish: %v", err)
	}
	found = nil
	for i := range runs {
		if runs[i].ID == id {
			found = &runs[i]
		}
	}
	if found == nil {
		t.Fatal("the finished run disappeared")
	}
	if found.OK {
		t.Error("the run failed and must be recorded as failed")
	}
	if !strings.Contains(found.Error, "i/o timeout") {
		t.Errorf("the reason must survive: %q", found.Error)
	}
	if found.FinishedAt == nil {
		t.Error("a finished run needs a finish time")
	}
	if found.Command != "run" || found.Host != "_test_host" {
		t.Errorf("run identity lost: %+v", found)
	}
}

func TestLiveRecentRunsAreNewestFirst(t *testing.T) {
	ctx, pool, db := liveDB(t)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM collector_runs WHERE host = '_test_host'`)
	})

	for _, cmd := range []string{"first", "second"} {
		id, err := db.StartRun(ctx, store.RunStart{Command: cmd, Host: "_test_host", Tier: "personal"})
		if err != nil {
			t.Fatalf("StartRun(%s): %v", cmd, err)
		}
		if err := db.FinishRun(ctx, id, store.RunFinish{OK: true}); err != nil {
			t.Fatalf("FinishRun(%s): %v", cmd, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	runs, err := db.RecentRuns(ctx, 10)
	if err != nil {
		t.Fatalf("RecentRuns: %v", err)
	}
	var seen []string
	for _, r := range runs {
		if r.Host == "_test_host" {
			seen = append(seen, r.Command)
		}
	}
	if len(seen) < 2 || seen[0] != "second" {
		t.Errorf("newest first expected, got %v", seen)
	}
}
