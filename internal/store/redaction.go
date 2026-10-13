package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// MaxRedactionAttempts caps how many times one photograph is scanned before it
// is left alone.
//
// The classifier learned this the hard way: a capture that can never succeed
// retried on every sweep, burning the work budget and sitting at the head of
// the queue. A photograph that has failed scanning five times needs a person,
// not a sixth attempt.
const MaxRedactionAttempts = 5

// RedactionTarget is a photograph nothing has looked at yet, together with the
// classifier's own reading of it.
//
// PeoplePresent is carried because it is the only independent reading of the
// same photograph available. It is used to contradict the detector, never to
// excuse it.
type RedactionTarget struct {
	MediaID       string
	ReportID      string
	ArchiveKey    string
	ContentType   string
	PeoplePresent bool
	Attempts      int
}

// RedactionResult is what happened, as it will be recorded.
type RedactionResult struct {
	RanAt     time.Time
	Tool      string
	OutputKey string
	Regions   int
	Published bool
	Reason    string
}

// MediaToRedact returns photographs with no redaction verdict, oldest first.
func (d *DB) MediaToRedact(ctx context.Context, limit int) ([]RedactionTarget, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT m.id::text, m.report_id::text, m.archive_key,
		       COALESCE(m.content_type, ''),
		       COALESCE((SELECT c.people_present FROM classifications c
		                  WHERE c.report_id = m.report_id
		                    AND c.people_present IS NOT NULL
		                  ORDER BY c.created_at DESC LIMIT 1), false),
		       m.redaction_attempts
		FROM report_media m
		WHERE m.redaction_verdict IS NULL
		  AND m.redaction_attempts < $1
		ORDER BY m.created_at
		LIMIT $2`, MaxRedactionAttempts, limit)
	if err != nil {
		return nil, fmt.Errorf("list media to redact: %w", err)
	}
	defer rows.Close()

	var out []RedactionTarget
	for rows.Next() {
		var t RedactionTarget
		if err := rows.Scan(&t.MediaID, &t.ReportID, &t.ArchiveKey,
			&t.ContentType, &t.PeoplePresent, &t.Attempts); err != nil {
			return nil, fmt.Errorf("scan media to redact: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read media to redact: %w", err)
	}
	return out, nil
}

// ErrMediaGone reports that the photograph was deleted while being scanned.
var ErrMediaGone = errors.New("media row no longer exists")

// RecordRedaction writes what the scan did. A held verdict stores no output
// key: the database refuses a published verdict without one, so the two cases
// cannot be confused by a later reader.
func (d *DB) RecordRedaction(ctx context.Context, mediaID string, r RedactionResult) error {
	verdict := "held"
	var outputKey *string
	if r.Published {
		verdict = "published"
		if r.OutputKey == "" {
			return fmt.Errorf("record redaction for %s: published with no output object", mediaID)
		}
		outputKey = &r.OutputKey
	}

	tag, err := d.pool.Exec(ctx, `
		UPDATE report_media SET
		  redaction_ran_at     = $2,
		  redaction_tool       = $3,
		  redaction_output_key = $4,
		  redaction_regions    = $5,
		  redaction_verdict    = $6,
		  redaction_reason     = NULLIF($7, ''),
		  redaction_attempts   = redaction_attempts + 1
		WHERE id = $1::uuid`,
		mediaID, r.RanAt.UTC(), r.Tool, outputKey, r.Regions, verdict, r.Reason)
	if err != nil {
		return fmt.Errorf("record redaction for %s: %w", mediaID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("record redaction for %s: %w", mediaID, ErrMediaGone)
	}
	return nil
}

// BumpRedactionAttempt counts a scan that could not be carried out at all —
// the photograph would not download, the cascade would not load. The verdict
// is left unset so the sweep tries again, up to MaxRedactionAttempts.
func (d *DB) BumpRedactionAttempt(ctx context.Context, mediaID string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE report_media SET redaction_attempts = redaction_attempts + 1
		 WHERE id = $1::uuid`, mediaID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("bump redaction attempt for %s: %w", mediaID, err)
	}
	return nil
}
