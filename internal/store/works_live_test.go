package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vinit-churi/tracesarkar/internal/config"
	"github.com/vinit-churi/tracesarkar/internal/ingest"
	"github.com/vinit-churi/tracesarkar/internal/sources"
	"github.com/vinit-churi/tracesarkar/internal/store"
	"github.com/vinit-churi/tracesarkar/internal/works"
)

func liveDB(t *testing.T) (context.Context, *pgxpool.Pool, *store.DB) {
	t.Helper()
	if os.Getenv("TRACESARKAR_LIVE") != "1" {
		t.Skip("set TRACESARKAR_LIVE=1 to run live database tests")
	}
	cfg, err := config.Load("../../.env")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	pool, err := store.Connect(ctx, cfg.Postgres.URL, cfg.Postgres.CAPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if _, err := store.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db := store.NewDB(pool)

	reg, err := sources.Load("../../data/sources.yaml")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := db.SyncRegister(ctx, reg); err != nil {
		t.Fatalf("sync register: %v", err)
	}
	return ctx, pool, db
}

// testEndpoint gives each run its own endpoint name so live runs stay isolated.
func testEndpoint(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("_test_%d", time.Now().UnixNano())
}

func TestLiveSnapshotRoundTrip(t *testing.T) {
	ctx, pool, db := liveDB(t)
	endpoint := testEndpoint(t)
	const source = "bmc_roads_api"

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM work_changes WHERE work_id IN
			(SELECT id FROM works WHERE source_id = $1 AND endpoint = $2)`, source, endpoint)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM works WHERE source_id = $1 AND endpoint = $2`, source, endpoint)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM fetch_log WHERE source_id = $1 AND endpoint = $2`, source, endpoint)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM raw_documents WHERE source_id = $1 AND endpoint = $2`, source, endpoint)
	})

	// No document yet.
	sha, err := db.LastDocumentSHA(ctx, source, endpoint)
	if err != nil {
		t.Fatalf("LastDocumentSHA on empty: %v", err)
	}
	if sha != "" {
		t.Fatalf("expected no document, got %q", sha)
	}

	const firstSHA = "1111111111111111111111111111111111111111111111111111111111111111"
	docID, err := db.SaveRawDocument(ctx, ingest.RawDocument{
		SourceID: source, Endpoint: endpoint, URL: "https://example.test/a",
		ArchiveKey: "archive/test/a.json", SHA256: firstSHA, ContentType: "application/json",
		Bytes: 10, RetrievedAt: time.Now().UTC(), CapturedBy: "ingester:test",
	})
	if err != nil {
		t.Fatalf("SaveRawDocument: %v", err)
	}
	if docID == "" {
		t.Fatal("expected a document id")
	}

	if sha, err = db.LastDocumentSHA(ctx, source, endpoint); err != nil || sha != firstSHA {
		t.Fatalf("LastDocumentSHA: got %q err=%v", sha, err)
	}

	if err := db.RecordFetch(ctx, ingest.FetchRecord{
		SourceID: source, Endpoint: endpoint, RequestedAt: time.Now().UTC(),
		StatusCode: 200, Bytes: 10, SHA256: firstSHA, DocumentID: docID, DurationMS: 42,
	}); err != nil {
		t.Fatalf("RecordFetch: %v", err)
	}

	// Apply a first snapshot: one record, recorded as added.
	rec := works.Record{NaturalKey: "Road A|R/S|X", Fields: map[string]any{"status": "In Progress", "ward": "R/S"}}
	if err := db.ApplySnapshot(ctx, ingest.SnapshotWrite{
		SourceID: source, Endpoint: endpoint, DocumentID: docID, ObservedAt: time.Now().UTC(),
		Records: []works.Record{rec},
		Changes: []works.Change{{Kind: works.KindAdded, NaturalKey: rec.NaturalKey}},
	}); err != nil {
		t.Fatalf("ApplySnapshot: %v", err)
	}

	loaded, err := db.LoadWorks(ctx, source, endpoint)
	if err != nil {
		t.Fatalf("LoadWorks: %v", err)
	}
	if len(loaded) != 1 || loaded[rec.NaturalKey].Fields["status"] != "In Progress" {
		t.Fatalf("round trip lost data: %+v", loaded)
	}

	// Second snapshot: the status changes.
	const secondSHA = "2222222222222222222222222222222222222222222222222222222222222222"
	docID2, err := db.SaveRawDocument(ctx, ingest.RawDocument{
		SourceID: source, Endpoint: endpoint, URL: "https://example.test/a",
		ArchiveKey: "archive/test/b.json", SHA256: secondSHA, ContentType: "application/json",
		Bytes: 11, RetrievedAt: time.Now().UTC(), CapturedBy: "ingester:test",
	})
	if err != nil {
		t.Fatalf("SaveRawDocument (second): %v", err)
	}

	updated := works.Record{NaturalKey: rec.NaturalKey, Fields: map[string]any{"status": "Completed", "ward": "R/S"}}
	observed := time.Now().UTC()
	if err := db.ApplySnapshot(ctx, ingest.SnapshotWrite{
		SourceID: source, Endpoint: endpoint, DocumentID: docID2, ObservedAt: observed,
		Records: []works.Record{updated},
		Changes: []works.Change{{
			Kind: works.KindChanged, NaturalKey: rec.NaturalKey,
			Field: "status", Old: "In Progress", New: "Completed",
		}},
	}); err != nil {
		t.Fatalf("ApplySnapshot (second): %v", err)
	}

	changes, err := db.ChangesSince(ctx, observed.Add(-time.Minute), 50)
	if err != nil {
		t.Fatalf("ChangesSince: %v", err)
	}
	var found bool
	for _, c := range changes {
		if c.Endpoint == endpoint && c.Kind == works.KindChanged && c.Field == "status" {
			found = true
			if c.Old == "" || c.New == "" {
				t.Errorf("both values must be stored: %+v", c)
			}
		}
	}
	if !found {
		t.Errorf("the status change was not recorded; got %d changes", len(changes))
	}

	loaded, err = db.LoadWorks(ctx, source, endpoint)
	if err != nil {
		t.Fatalf("LoadWorks (second): %v", err)
	}
	if loaded[rec.NaturalKey].Fields["status"] != "Completed" {
		t.Errorf("current state not updated: %+v", loaded[rec.NaturalKey].Fields)
	}
}

func TestLiveSaveRawDocumentIsIdempotentOnIdenticalBytes(t *testing.T) {
	ctx, pool, db := liveDB(t)
	endpoint := testEndpoint(t)
	const source = "bmc_roads_api"
	const sha = "3333333333333333333333333333333333333333333333333333333333333333"

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM raw_documents WHERE sha256 = decode($1,'hex')`, sha)
	})

	doc := ingest.RawDocument{
		SourceID: source, Endpoint: endpoint, URL: "https://example.test/dup",
		ArchiveKey: "archive/test/dup.json", SHA256: sha, ContentType: "application/json",
		Bytes: 5, RetrievedAt: time.Now().UTC(), CapturedBy: "ingester:test",
	}
	first, err := db.SaveRawDocument(ctx, doc)
	if err != nil {
		t.Fatalf("first save: %v", err)
	}
	second, err := db.SaveRawDocument(ctx, doc)
	if err != nil {
		t.Fatalf("second save: %v", err)
	}
	if first != second {
		t.Errorf("identical bytes must map to one document: %s vs %s", first, second)
	}
}
