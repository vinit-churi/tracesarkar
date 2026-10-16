package archive_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/config"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// A row that cites an archive key the bucket does not hold is worse than a row
// with no key at all: it claims an evidentiary record exists and nobody finds
// out otherwise until somebody has to defend the claim. Every key the database
// points at must actually be in the archive.
func TestLiveEveryCitedArchiveKeyExists(t *testing.T) {
	blobs := liveClient(t)

	cfg, err := config.Load("../../.env")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := store.Connect(ctx, cfg.Postgres.URL, cfg.Postgres.CAPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, `
		SELECT DISTINCT archive_key FROM authority_departments
		 WHERE archive_key IS NOT NULL AND archive_key <> ''
		UNION
		SELECT DISTINCT archive_key FROM raw_documents
		 WHERE archive_key IS NOT NULL AND archive_key <> ''
		   AND source_id = 'bmc_portal_rti_manuals'`)
	if err != nil {
		t.Fatalf("read keys: %v", err)
	}
	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			t.Fatalf("scan: %v", err)
		}
		keys = append(keys, k)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		t.Skip("no archive keys are cited yet")
	}

	var missing []string
	for _, k := range keys {
		ok, err := blobs.Exists(ctx, k)
		if err != nil {
			t.Fatalf("head %s: %v", k, err)
		}
		if !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d of %d cited archive key(s) are not in the bucket:", len(missing), len(keys))
		for _, k := range missing {
			t.Errorf("  %s", k)
		}
	}
	t.Logf("%d cited archive key(s) checked", len(keys))
}
