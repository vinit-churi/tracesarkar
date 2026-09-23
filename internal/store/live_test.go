package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/config"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

func TestLiveMigrateIsIdempotent(t *testing.T) {
	if os.Getenv("TRACESARKAR_LIVE") != "1" {
		t.Skip("set TRACESARKAR_LIVE=1 to run live database tests")
	}
	cfg, err := config.Load("../../.env")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pool, err := store.Connect(ctx, cfg.Postgres.URL, cfg.Postgres.CAPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if _, err := store.Migrate(ctx, pool); err != nil {
		t.Fatalf("first migrate: %v", err)
	}

	second, err := store.Migrate(ctx, pool)
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if len(second) != 0 {
		t.Errorf("second run applied %d migrations, want 0", len(second))
	}

	// Every Phase 0 table should now exist.
	for _, table := range []string{
		"sources", "raw_documents", "fetch_log", "works", "work_changes",
		"gr_items", "court_items", "legal_constants", "schema_migrations",
	} {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables
			                WHERE table_schema = 'public' AND table_name = $1)`, table).Scan(&exists)
		if err != nil {
			t.Fatalf("check %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %s missing after migrate", table)
		}
	}
}
