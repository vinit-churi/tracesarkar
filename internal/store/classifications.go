package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vinit-churi/tracesarkar/internal/classify"
)

// NewClassification is one attempt at classifying a capture, successful or not.
type NewClassification struct {
	ReportID string
	Decision classify.Decision
	Model    string
	Prompt   string
	Latency  time.Duration
	Err      string
}

// SaveClassification records an attempt.
//
// Failures are stored too: a classification that silently never ran looks
// exactly like one that ran and found nothing, and the two need very
// different responses.
func (d *DB) SaveClassification(ctx context.Context, in NewClassification) error {
	dec := in.Decision
	_, err := d.pool.Exec(ctx, `
		INSERT INTO classifications (
		  id, report_id, category, subcategory, severity, hazard_to_life,
		  water_present, people_present, image_quality, is_civic_issue,
		  confidence, rationale, outcome, overridden, override_reason,
		  redaction_required, rejection_reason, model, prompt_version,
		  latency_ms, error)
		VALUES (gen_random_uuid(), $1::uuid, NULLIF($2,''), NULLIF($3,''),
		        NULLIF($4,''), $5, $6, $7, NULLIF($8,''), $9, $10::numeric,
		        NULLIF($11,''), $12, $13, NULLIF($14,''), $15, NULLIF($16,''),
		        $17, $18, $19, NULLIF($20,''))`,
		in.ReportID, dec.Category, dec.Subcategory, dec.Severity, dec.HazardToLife,
		dec.WaterPresent, dec.PeoplePresent, dec.ImageQuality, dec.IsCivicIssue,
		dec.Confidence, dec.Rationale, string(dec.Outcome), dec.Overridden,
		dec.OverrideReason, dec.RedactionRequired, dec.RejectionReason,
		in.Model, in.Prompt, in.Latency.Milliseconds(), in.Err)
	if err != nil {
		return fmt.Errorf("save classification for %s: %w", in.ReportID, err)
	}
	return nil
}

// LatestClassification returns the most recent successful classification of a
// capture, or false when there is none.
func (d *DB) LatestClassification(ctx context.Context, reportID string) (classify.Decision, bool, error) {
	var dec classify.Decision
	err := d.pool.QueryRow(ctx, `
		SELECT COALESCE(category,''), COALESCE(subcategory,''), COALESCE(severity,''),
		       COALESCE(hazard_to_life,false), COALESCE(water_present,false),
		       COALESCE(people_present,false), COALESCE(image_quality,''),
		       COALESCE(is_civic_issue,false), COALESCE(confidence,0)::float8,
		       COALESCE(rationale,''), outcome, overridden,
		       COALESCE(override_reason,''), redaction_required,
		       COALESCE(rejection_reason,'')
		  FROM classifications
		 WHERE report_id = $1::uuid AND error IS NULL
		 ORDER BY created_at DESC LIMIT 1`, reportID).Scan(
		&dec.Category, &dec.Subcategory, &dec.Severity, &dec.HazardToLife,
		&dec.WaterPresent, &dec.PeoplePresent, &dec.ImageQuality,
		&dec.IsCivicIssue, &dec.Confidence, &dec.Rationale,
		(*string)(&dec.Outcome), &dec.Overridden, &dec.OverrideReason,
		&dec.RedactionRequired, &dec.RejectionReason)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return classify.Decision{}, false, nil
		}
		return classify.Decision{}, false, fmt.Errorf("read classification: %w", err)
	}
	return dec, true, nil
}

// MaxClassifyAttempts is how often one capture is retried before it is left
// alone.
//
// Some photographs can never be classified — corrupt bytes, a format the
// provider rejects, a frame the model refuses. Retrying those without end costs
// a paid vision call per sweep forever, and because the queue is oldest-first
// with a limit, they sit at its head and crowd out captures taken after them.
// Hard rule 7 requires every stage to be retryable; it does not require them to
// be retried without end.
//
// A retired capture is not lost. Its attempts are all recorded with their
// errors, and raising this number brings it straight back into the queue.
const MaxClassifyAttempts = 5

// UnclassifiedReports returns captures that have never been classified
// successfully and have not exhausted their attempts, oldest first — a capture
// is never lost, so the queue drains forwards.
func (d *DB) UnclassifiedReports(ctx context.Context, limit int) ([]PendingCapture, error) {
	if limit <= 0 || limit > 200 {
		limit = 25
	}
	rows, err := d.pool.Query(ctx, `
		SELECT r.id::text, m.archive_key, COALESCE(j.ward, '')
		  FROM reports r
		  JOIN LATERAL (
		    SELECT archive_key FROM report_media
		     WHERE report_id = r.id ORDER BY created_at LIMIT 1
		  ) m ON true
		  LEFT JOIN report_jurisdiction j ON j.report_id = r.id
		 WHERE NOT EXISTS (
		    SELECT 1 FROM classifications c
		     WHERE c.report_id = r.id AND c.error IS NULL)
		   AND (SELECT count(*) FROM classifications c
		         WHERE c.report_id = r.id AND c.error IS NOT NULL) < $2
		 ORDER BY r.created_at
		 LIMIT $1`, limit, MaxClassifyAttempts)
	if err != nil {
		return nil, fmt.Errorf("find unclassified reports: %w", err)
	}
	defer rows.Close()

	var out []PendingCapture
	for rows.Next() {
		var p PendingCapture
		if err := rows.Scan(&p.ReportID, &p.ArchiveKey, &p.Ward); err != nil {
			return nil, fmt.Errorf("scan pending capture: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// PendingCapture is a stored photograph waiting to be classified.
type PendingCapture struct {
	ReportID   string
	ArchiveKey string
	// Ward, empty when jurisdiction has not resolved yet. Empty means not
	// covered, because a capture whose ward is unknown cannot be known to
	// have a department behind it.
	Ward string
}

// CoverageByWard lists, per ward, the categories that have a department behind
// them — which is the set the platform can actually route *there*.
//
// Per ward because a department recorded for R/Central says nothing about
// R/North. Asking only "is this category covered anywhere" accepted a waste
// report in Dahisar on the strength of a Borivali department, which promises
// the citizen a desk that does not exist for them.
func (d *DB) CoverageByWard(ctx context.Context) (map[string][]string, error) {
	rows, err := d.pool.Query(ctx,
		`SELECT DISTINCT ward, category FROM authority_departments WHERE ward <> ''`)
	if err != nil {
		return nil, fmt.Errorf("coverage by ward: %w", err)
	}
	defer rows.Close()

	out := map[string][]string{}
	for rows.Next() {
		var ward, category string
		if err := rows.Scan(&ward, &category); err != nil {
			return nil, fmt.Errorf("scan coverage: %w", err)
		}
		out[ward] = append(out[ward], category)
	}
	return out, rows.Err()
}

// CoveredCategories lists the issue categories that have a department behind
// them anywhere. Kept for reporting on how far coverage has got; routing uses
// CoverageByWard, because "covered somewhere" is not "covered here".
func (d *DB) CoveredCategories(ctx context.Context) ([]string, error) {
	rows, err := d.pool.Query(ctx, `SELECT DISTINCT category FROM authority_departments`)
	if err != nil {
		return nil, fmt.Errorf("covered categories: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
