package store

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vinit-churi/tracesarkar/internal/ingest"
	"github.com/vinit-churi/tracesarkar/internal/sources"
	"github.com/vinit-churi/tracesarkar/internal/watch"
	"github.com/vinit-churi/tracesarkar/internal/works"
)

// DB implements the ingest.Store interface against PostgreSQL. SQL is written
// by hand here, as the conventions require.
type DB struct {
	pool *pgxpool.Pool
}

// NewDB wraps a pool.
func NewDB(pool *pgxpool.Pool) *DB { return &DB{pool: pool} }

// SyncRegister upserts the register into the sources table so that foreign keys
// and operator queries work against the same list the code enforces.
func (d *DB) SyncRegister(ctx context.Context, reg *sources.Register) error {
	for _, src := range reg.All() {
		_, err := d.pool.Exec(ctx, `
			INSERT INTO sources (id, name, url, category, acquisition, cadence, licence,
			                     terms_reviewed_on, terms_reviewed_by, robots_ok, status,
			                     blocklist, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8::date,$9,$10,$11,$12,now())
			ON CONFLICT (id) DO UPDATE SET
			  name = EXCLUDED.name, url = EXCLUDED.url, category = EXCLUDED.category,
			  acquisition = EXCLUDED.acquisition, cadence = EXCLUDED.cadence,
			  licence = EXCLUDED.licence, terms_reviewed_on = EXCLUDED.terms_reviewed_on,
			  terms_reviewed_by = EXCLUDED.terms_reviewed_by, robots_ok = EXCLUDED.robots_ok,
			  status = EXCLUDED.status, blocklist = EXCLUDED.blocklist, updated_at = now()`,
			src.ID, src.Name, nullString(src.URL), src.Category, src.Acquisition,
			nullString(src.Cadence), nullString(src.Licence), src.TermsReviewedOn,
			src.TermsReviewedBy, src.RobotsOK, src.Status, blocklistOrEmpty(src.Blocklist))
		if err != nil {
			return fmt.Errorf("sync source %s: %w", src.ID, err)
		}
	}
	return nil
}

// LastDocumentSHA returns the hash of the most recent archived body for an
// endpoint, or "" when there is none.
func (d *DB) LastDocumentSHA(ctx context.Context, sourceID, endpoint string) (string, error) {
	var sum []byte
	err := d.pool.QueryRow(ctx, `
		SELECT sha256 FROM raw_documents
		WHERE source_id = $1 AND endpoint = $2 AND parsed_at IS NOT NULL
		ORDER BY retrieved_at DESC
		LIMIT 1`, sourceID, endpoint).Scan(&sum)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("last document: %w", err)
	}
	return hex.EncodeToString(sum), nil
}

// RecordFetch writes one fetch_log row.
func (d *DB) RecordFetch(ctx context.Context, rec ingest.FetchRecord) error {
	var sum []byte
	if rec.SHA256 != "" {
		decoded, err := hex.DecodeString(rec.SHA256)
		if err != nil {
			return fmt.Errorf("decode sha256: %w", err)
		}
		sum = decoded
	}
	_, err := d.pool.Exec(ctx, `
		INSERT INTO fetch_log (id, source_id, endpoint, requested_at, status_code, bytes,
		                       sha256, raw_document_id, unchanged, duration_ms, error)
		VALUES (gen_random_uuid(), $1,$2,$3,
		        NULLIF($4,0), NULLIF($5,0)::bigint, $6, NULLIF($7,'')::uuid, $8,
		        NULLIF($9,0), NULLIF($10,''))`,
		rec.SourceID, rec.Endpoint, rec.RequestedAt, rec.StatusCode, rec.Bytes,
		sum, rec.DocumentID, rec.Unchanged, rec.DurationMS, rec.Error)
	if err != nil {
		return fmt.Errorf("record fetch: %w", err)
	}
	return nil
}

// SaveRawDocument records an archived artefact and returns its id. A re-fetch of
// identical bytes returns the existing id.
func (d *DB) SaveRawDocument(ctx context.Context, doc ingest.RawDocument) (string, error) {
	sum, err := hex.DecodeString(doc.SHA256)
	if err != nil {
		return "", fmt.Errorf("decode sha256: %w", err)
	}
	var id string
	err = d.pool.QueryRow(ctx, `
		INSERT INTO raw_documents (id, source_id, endpoint, url, archive_key, sha256,
		                           content_type, bytes, retrieved_at, captured_by)
		VALUES (gen_random_uuid(), $1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (sha256) DO UPDATE SET retrieved_at = raw_documents.retrieved_at
		RETURNING id::text`,
		doc.SourceID, doc.Endpoint, nullString(doc.URL), doc.ArchiveKey, sum,
		nullString(doc.ContentType), doc.Bytes, doc.RetrievedAt, doc.CapturedBy).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("save raw document: %w", err)
	}
	return id, nil
}

// MarkDocumentParsed records a successful parse. Documents without it are
// retried on the next run, so an archived body is never silently skipped.
func (d *DB) MarkDocumentParsed(ctx context.Context, documentID string) error {
	_, err := d.pool.Exec(ctx, `
		UPDATE raw_documents SET parsed_at = now(), parse_status = 'ok'
		WHERE id = $1::uuid`, documentID)
	if err != nil {
		return fmt.Errorf("mark document parsed: %w", err)
	}
	return nil
}

// LoadWorks returns the current state of an endpoint's records.
func (d *DB) LoadWorks(ctx context.Context, sourceID, endpoint string) (map[string]works.Record, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT natural_key, current, vanished_at IS NOT NULL
		FROM works WHERE source_id = $1 AND endpoint = $2`, sourceID, endpoint)
	if err != nil {
		return nil, fmt.Errorf("load works: %w", err)
	}
	defer rows.Close()

	out := map[string]works.Record{}
	for rows.Next() {
		var key string
		var raw []byte
		var vanished bool
		if err := rows.Scan(&key, &raw, &vanished); err != nil {
			return nil, fmt.Errorf("scan works: %w", err)
		}
		fields := map[string]any{}
		if err := json.Unmarshal(raw, &fields); err != nil {
			return nil, fmt.Errorf("decode works row %s: %w", key, err)
		}
		out[key] = works.Record{NaturalKey: key, Fields: fields, Vanished: vanished}
	}
	return out, rows.Err()
}

// ApplySnapshot writes the parsed records and their changes in one transaction.
func (d *DB) ApplySnapshot(ctx context.Context, in ingest.SnapshotWrite) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	seen := make(map[string]bool, len(in.Records))
	for _, rec := range in.Records {
		encoded, err := json.Marshal(rec.Fields)
		if err != nil {
			return fmt.Errorf("encode %s: %w", rec.NaturalKey, err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO works (id, source_id, endpoint, natural_key, current,
			                   first_seen_at, last_seen_at, vanished_at, raw_document_id)
			VALUES (gen_random_uuid(), $1,$2,$3,$4,$5,$5,NULL,$6::uuid)
			ON CONFLICT (source_id, endpoint, natural_key) DO UPDATE SET
			  current = EXCLUDED.current,
			  last_seen_at = EXCLUDED.last_seen_at,
			  vanished_at = NULL,
			  raw_document_id = EXCLUDED.raw_document_id`,
			in.SourceID, in.Endpoint, rec.NaturalKey, encoded, in.ObservedAt, in.DocumentID)
		if err != nil {
			return fmt.Errorf("upsert work %s: %w", rec.NaturalKey, err)
		}
		seen[rec.NaturalKey] = true
	}

	for _, change := range in.Changes {
		if change.Kind == works.KindVanished {
			if _, err := tx.Exec(ctx, `
				UPDATE works SET vanished_at = $4
				WHERE source_id = $1 AND endpoint = $2 AND natural_key = $3 AND vanished_at IS NULL`,
				in.SourceID, in.Endpoint, change.NaturalKey, in.ObservedAt); err != nil {
				return fmt.Errorf("mark vanished %s: %w", change.NaturalKey, err)
			}
		}

		oldJSON, err := marshalNullable(change.Old)
		if err != nil {
			return err
		}
		newJSON, err := marshalNullable(change.New)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO work_changes (id, work_id, kind, field, old_value, new_value,
			                          prev_raw_document_id, raw_document_id, observed_at)
			SELECT gen_random_uuid(), w.id, $4, NULLIF($5,''), $6, $7, w.raw_document_id, $8::uuid, $9
			FROM works w
			WHERE w.source_id = $1 AND w.endpoint = $2 AND w.natural_key = $3`,
			in.SourceID, in.Endpoint, change.NaturalKey, change.Kind, change.Field,
			oldJSON, newJSON, in.DocumentID, in.ObservedAt)
		if err != nil {
			return fmt.Errorf("record change for %s: %w", change.NaturalKey, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit snapshot: %w", err)
	}
	return nil
}

// MarkSourceRun updates the register table after a run.
func (d *DB) MarkSourceRun(ctx context.Context, sourceID string, at time.Time, runErr string) error {
	_, err := d.pool.Exec(ctx, `
		UPDATE sources
		SET last_success_at = CASE WHEN $3 = '' THEN $2 ELSE last_success_at END,
		    last_error = NULLIF($3,'')
		WHERE id = $1`, sourceID, at, runErr)
	if err != nil {
		return fmt.Errorf("mark source run: %w", err)
	}
	return nil
}

// ChangesSince lists recent changes, newest first.
func (d *DB) ChangesSince(ctx context.Context, since time.Time, limit int) ([]ChangeRow, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT w.source_id, w.endpoint, w.natural_key, c.kind, COALESCE(c.field,''),
		       COALESCE(c.old_value::text,''), COALESCE(c.new_value::text,''), c.observed_at
		FROM work_changes c JOIN works w ON w.id = c.work_id
		WHERE c.observed_at >= $1
		ORDER BY c.observed_at DESC, w.natural_key
		LIMIT $2`, since, limit)
	if err != nil {
		return nil, fmt.Errorf("changes since: %w", err)
	}
	defer rows.Close()

	var out []ChangeRow
	for rows.Next() {
		var r ChangeRow
		if err := rows.Scan(&r.SourceID, &r.Endpoint, &r.NaturalKey, &r.Kind,
			&r.Field, &r.Old, &r.New, &r.ObservedAt); err != nil {
			return nil, fmt.Errorf("scan change: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ChangeRow is one row of the change log for display.
type ChangeRow struct {
	SourceID   string
	Endpoint   string
	NaturalKey string
	Kind       string
	Field      string
	Old        string
	New        string
	ObservedAt time.Time
}

// SnapshotStatus summarises collection health per source and endpoint.
type SnapshotStatus struct {
	SourceID    string
	Endpoint    string
	LastFetchAt *time.Time
	LastChange  *time.Time
	Records     int
	Attempts    int
	Failures    int
}

// Status reports per-endpoint collection health.
func (d *DB) Status(ctx context.Context) ([]SnapshotStatus, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT f.source_id, f.endpoint,
		       max(f.requested_at) AS last_fetch,
		       count(*) AS attempts,
		       count(*) FILTER (WHERE f.error IS NOT NULL) AS failures,
		       (SELECT count(*) FROM works w WHERE w.source_id = f.source_id AND w.endpoint = f.endpoint
		                                       AND w.vanished_at IS NULL) AS records,
		       (SELECT max(c.observed_at) FROM work_changes c
		          JOIN works w2 ON w2.id = c.work_id
		         WHERE w2.source_id = f.source_id AND w2.endpoint = f.endpoint) AS last_change
		FROM fetch_log f
		GROUP BY f.source_id, f.endpoint
		ORDER BY f.source_id, f.endpoint`)
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}
	defer rows.Close()

	var out []SnapshotStatus
	for rows.Next() {
		var s SnapshotStatus
		if err := rows.Scan(&s.SourceID, &s.Endpoint, &s.LastFetchAt, &s.Attempts,
			&s.Failures, &s.Records, &s.LastChange); err != nil {
			return nil, fmt.Errorf("scan status: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func marshalNullable(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encode value: %w", err)
	}
	return b, nil
}

// blocklistOrEmpty keeps the column non-null for sources with nothing to strip.
func blocklistOrEmpty(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// KnownGRs reports which of these sanketanks are already recorded.
func (d *DB) KnownGRs(ctx context.Context, sanketanks []string) (map[string]bool, error) {
	out := map[string]bool{}
	if len(sanketanks) == 0 {
		return out, nil
	}
	rows, err := d.pool.Query(ctx, `SELECT sanketank FROM gr_items WHERE sanketank = ANY($1)`, sanketanks)
	if err != nil {
		return nil, fmt.Errorf("known resolutions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, fmt.Errorf("scan resolution: %w", err)
		}
		out[s] = true
	}
	return out, rows.Err()
}

// SaveGR records one Government Resolution.
func (d *DB) SaveGR(ctx context.Context, rec watch.GRRecord) error {
	keywords := rec.Keywords
	if keywords == nil {
		keywords = []string{}
	}
	_, err := d.pool.Exec(ctx, `
		INSERT INTO gr_items (sanketank, issued_on, title, raw_document_id, keywords)
		VALUES ($1, $2::date, $3, NULLIF($4,'')::uuid, $5)
		ON CONFLICT (sanketank) DO UPDATE SET
		  issued_on = COALESCE(EXCLUDED.issued_on, gr_items.issued_on),
		  title = COALESCE(EXCLUDED.title, gr_items.title),
		  raw_document_id = COALESCE(EXCLUDED.raw_document_id, gr_items.raw_document_id),
		  keywords = EXCLUDED.keywords`,
		rec.Sanketank, rec.IssuedOn, nullString(rec.Title), rec.DocumentID, keywords)
	if err != nil {
		return fmt.Errorf("save resolution %s: %w", rec.Sanketank, err)
	}
	return nil
}

// GRHits lists recorded resolutions that mention a watched term.
func (d *DB) GRHits(ctx context.Context, limit int) ([]GRHitRow, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT sanketank, COALESCE(title,''), issued_on, keywords
		FROM gr_items
		WHERE cardinality(keywords) > 0
		ORDER BY issued_on DESC NULLS LAST, sanketank DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("resolution hits: %w", err)
	}
	defer rows.Close()

	var out []GRHitRow
	for rows.Next() {
		var r GRHitRow
		if err := rows.Scan(&r.Sanketank, &r.Title, &r.IssuedOn, &r.Keywords); err != nil {
			return nil, fmt.Errorf("scan resolution hit: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GRHitRow is a resolution worth reading.
type GRHitRow struct {
	Sanketank string
	Title     string
	IssuedOn  *time.Time
	Keywords  []string
}
