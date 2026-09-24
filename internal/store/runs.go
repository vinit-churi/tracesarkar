package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// RunStart identifies a collector invocation.
type RunStart struct {
	Command string
	Host    string
	Tier    string
	Version string
}

// RunFinish is how it ended.
type RunFinish struct {
	OK     bool
	Error  string
	Detail map[string]any
}

// RunRow is a recorded run.
type RunRow struct {
	ID         string
	Command    string
	Host       string
	Tier       string
	StartedAt  time.Time
	FinishedAt *time.Time
	OK         bool
	Error      string
	Detail     map[string]any
}

// StartRun records that a run began, before anything can go wrong. A run with
// no finish time is one that never returned.
func (d *DB) StartRun(ctx context.Context, in RunStart) (string, error) {
	var id string
	err := d.pool.QueryRow(ctx, `
		INSERT INTO collector_runs (id, command, host, tier, version)
		VALUES (gen_random_uuid(), $1, $2, $3, NULLIF($4,''))
		RETURNING id::text`, in.Command, in.Host, in.Tier, in.Version).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("start run: %w", err)
	}
	return id, nil
}

// FinishRun records how a run ended.
func (d *DB) FinishRun(ctx context.Context, id string, in RunFinish) error {
	detail := in.Detail
	if detail == nil {
		detail = map[string]any{}
	}
	encoded, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("encode run detail: %w", err)
	}
	_, err = d.pool.Exec(ctx, `
		UPDATE collector_runs
		SET finished_at = now(), ok = $2, error = NULLIF($3,''), detail = $4
		WHERE id = $1::uuid`, id, in.OK, in.Error, encoded)
	if err != nil {
		return fmt.Errorf("finish run: %w", err)
	}
	return nil
}

// RecentRuns lists runs, newest first.
func (d *DB) RecentRuns(ctx context.Context, limit int) ([]RunRow, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT id::text, command, host, tier, started_at, finished_at,
		       coalesce(ok,false), coalesce(error,''), detail
		FROM collector_runs
		ORDER BY started_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("recent runs: %w", err)
	}
	defer rows.Close()

	var out []RunRow
	for rows.Next() {
		var r RunRow
		var detail []byte
		if err := rows.Scan(&r.ID, &r.Command, &r.Host, &r.Tier,
			&r.StartedAt, &r.FinishedAt, &r.OK, &r.Error, &detail); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}
		if len(detail) > 0 {
			_ = json.Unmarshal(detail, &r.Detail)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
