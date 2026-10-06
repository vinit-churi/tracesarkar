package store

import (
	"context"
	"fmt"
)

// EvalEntry is one labelled capture, as the evaluation set sees it.
//
// Ground truth is the single subcategory a person chose. Category and hazard
// are derived from the taxonomy by the classify package rather than stored
// again here: three fields that can disagree eventually will, and a wrong
// hazard flag silently moves the recall figure the model is judged against.
type EvalEntry struct {
	ReportID   string
	ArchiveKey string
	Label      string
	FrameType  string
	Conditions []string
	Notes      string
	Ward       string
}

// EvalSet returns every labelled capture that still has a photograph, oldest
// first.
//
// Oldest first so a run is reproducible: the same labels in the same order give
// the same report, and two runs can be compared line by line.
func (d *DB) EvalSet(ctx context.Context, limit int) ([]EvalEntry, error) {
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}
	rows, err := d.pool.Query(ctx, `
		SELECT r.id::text, m.archive_key, l.label,
		       COALESCE(l.frame_type, ''), COALESCE(l.conditions, '{}'),
		       COALESCE(l.notes, ''),
		       COALESCE(NULLIF(l.ward_ground_truth, ''), j.ward, '')
		  FROM report_labels l
		  JOIN reports r ON r.id = l.report_id
		  JOIN LATERAL (
		    SELECT archive_key FROM report_media
		     WHERE report_id = r.id ORDER BY created_at LIMIT 1
		  ) m ON true
		  LEFT JOIN report_jurisdiction j ON j.report_id = r.id
		 WHERE l.label <> ''
		 ORDER BY r.captured_at, r.id
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("eval set: %w", err)
	}
	defer rows.Close()

	var out []EvalEntry
	for rows.Next() {
		var e EvalEntry
		if err := rows.Scan(&e.ReportID, &e.ArchiveKey, &e.Label,
			&e.FrameType, &e.Conditions, &e.Notes, &e.Ward); err != nil {
			return nil, fmt.Errorf("scan eval entry: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
