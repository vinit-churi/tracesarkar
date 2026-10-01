package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// PendingLabel is a capture waiting for a human judgement, with everything the
// labelling page needs to show it.
//
// A photograph without a label is a photograph. The eval set the classifier is
// measured against is made of these, one judgement at a time.
type PendingLabel struct {
	ReportID    string
	ArchiveKey  string
	ContentType string
	Ward        string
	CapturedAt  time.Time
}

// The classifier's own answer is deliberately absent. A labeller who can see
// what the model said is grading the model rather than judging the photograph,
// and the eval set stops being independent of the thing it measures. The same
// mistake was made once on the attribution review queue and removed there.

// UnlabelledReports returns captures with a photograph and no human label,
// oldest first.
//
// Oldest first for the same reason the classification queue is: a day of
// walking produces captures in the order they were taken, and labelling them
// in that order keeps the labeller's memory of the walk in step with what they
// are looking at.
func (d *DB) UnlabelledReports(ctx context.Context, limit int) ([]PendingLabel, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := d.pool.Query(ctx, `
		SELECT r.id::text, m.archive_key, m.content_type,
		       COALESCE(j.ward, ''), r.captured_at
		  FROM reports r
		  JOIN LATERAL (
		    SELECT archive_key, content_type FROM report_media
		     WHERE report_id = r.id ORDER BY created_at LIMIT 1
		  ) m ON true
		  LEFT JOIN report_jurisdiction j ON j.report_id = r.id
		 WHERE NOT EXISTS (
		    SELECT 1 FROM report_labels l WHERE l.report_id = r.id)
		 ORDER BY r.captured_at
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("find unlabelled reports: %w", err)
	}
	defer rows.Close()

	var out []PendingLabel
	for rows.Next() {
		var p PendingLabel
		if err := rows.Scan(&p.ReportID, &p.ArchiveKey, &p.ContentType,
			&p.Ward, &p.CapturedAt); err != nil {
			return nil, fmt.Errorf("scan pending label: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ErrNotFound is returned when a row the caller named does not exist, or
// belongs to someone else. The two are deliberately the same error: telling a
// caller that a report exists but is not theirs discloses that it exists.
var ErrNotFound = errors.New("not found")

// MediaFor returns the first photograph of a report, with the account that owns
// it so the caller can refuse to serve someone else's capture.
func (d *DB) MediaFor(ctx context.Context, reportID string) (key, contentType, owner string, err error) {
	err = d.pool.QueryRow(ctx, `
		SELECT m.archive_key, m.content_type, r.account_id::text
		  FROM reports r
		  JOIN report_media m ON m.report_id = r.id
		 WHERE r.id = $1::uuid
		 ORDER BY m.created_at
		 LIMIT 1`, reportID).Scan(&key, &contentType, &owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", ErrNotFound
	}
	if err != nil {
		return "", "", "", fmt.Errorf("media for report: %w", err)
	}
	return key, contentType, owner, nil
}
