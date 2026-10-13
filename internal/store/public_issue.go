package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/vinit-churi/tracesarkar/internal/publish"
)

// PublicEvent is one entry of an issue's history as a stranger may read it.
//
// It has no actor field. The actor on an issue event is the account that did
// the thing, and an account identifies a person — so rather than filtering it
// out at the edge, where one forgotten handler reinstates it, it is absent from
// the type the public path can hold.
type PublicEvent struct {
	Seq    int
	Kind   string
	At     time.Time
	Detail string
	Hash   string
}

// PublicIssueView is everything a public issue page is allowed to know.
//
// Lat and Lon are the exact centroid and are not for publication: they are
// here because the coarsening belongs at the edge, where publish.Coarsen can
// state the cell width alongside the point. Nothing else on this struct is
// personal: no reporter, no media key, no complaint reference.
type PublicIssueView struct {
	IssueID     string
	Authority   string
	Ward        string
	Category    string
	Subcategory string
	Status      string
	Sightings   int

	Lat, Lon float64

	FirstSeen  time.Time
	LastSeen   time.Time
	NotifiedAt *time.Time

	// Reference is the complaint number the authority issued. It is read so
	// that a handler can tell a lodged issue from an unlodged one without a
	// second query, and it is never rendered: it belongs to the person who
	// lodged it and leads back to their details on the authority's own portal.
	Reference string

	// DeadlineHours and DeadlineCitation travel together. A deadline shown
	// without the document that creates it is an assertion, so a missing
	// citation means no clock is shown at all.
	DeadlineHours    int
	DeadlineCitation string

	// Redaction is the record for the photograph that would be shown. Nil when
	// no capture of this issue has been through redaction.
	Redaction *publish.Redaction

	Events []PublicEvent
}

// PublicIssue returns what may be published about one issue.
//
// An unparseable id is reported as absent rather than as an error: a stranger
// typing a wrong permalink has not caused a fault, and a 500 would tell them
// their guess was differently wrong from a 404.
func (d *DB) PublicIssue(ctx context.Context, issueID string) (PublicIssueView, bool, error) {
	var v PublicIssueView
	err := d.pool.QueryRow(ctx, `
		SELECT i.id::text, COALESCE(i.authority,''), COALESCE(i.ward,''),
		       i.category, COALESCE(i.subcategory,''), i.status,
		       (SELECT count(*) FROM issue_reports r WHERE r.issue_id = i.id),
		       ST_Y(i.centroid::geometry), ST_X(i.centroid::geometry),
		       i.first_seen_at, i.last_seen_at, i.notified_at,
		       COALESCE(i.reference,'')
		  FROM issues i
		 WHERE i.id = $1::uuid`, issueID).
		Scan(&v.IssueID, &v.Authority, &v.Ward, &v.Category, &v.Subcategory,
			&v.Status, &v.Sightings, &v.Lat, &v.Lon,
			&v.FirstSeen, &v.LastSeen, &v.NotifiedAt, &v.Reference)
	switch {
	case errors.Is(err, pgx.ErrNoRows), err != nil && isBadUUID(err):
		return PublicIssueView{}, false, nil
	case err != nil:
		return PublicIssueView{}, false, fmt.Errorf("public issue: %w", err)
	}

	if err := d.attachRedaction(ctx, &v); err != nil {
		return PublicIssueView{}, false, err
	}
	if err := d.attachPublicEvents(ctx, &v); err != nil {
		return PublicIssueView{}, false, err
	}
	if err := d.attachDeadline(ctx, &v); err != nil {
		return PublicIssueView{}, false, err
	}
	return v, true, nil
}

// attachRedaction finds the earliest photograph of this issue that has been
// through redaction — the original sighting, which is the evidence of the
// problem rather than of its age.
//
// It reads every verdict, not only published ones, so the page can say why a
// photograph is missing instead of silently having none.
func (d *DB) attachRedaction(ctx context.Context, v *PublicIssueView) error {
	var r publish.Redaction
	var verdict string
	err := d.pool.QueryRow(ctx, `
		SELECT m.redaction_ran_at, COALESCE(m.redaction_tool,''),
		       COALESCE(m.redaction_output_key,''), COALESCE(m.redaction_regions,0),
		       m.redaction_verdict
		  FROM issue_reports ir
		  JOIN report_media m ON m.report_id = ir.report_id
		 WHERE ir.issue_id = $1::uuid AND m.redaction_verdict IS NOT NULL
		 ORDER BY m.created_at
		 LIMIT 1`, v.IssueID).
		Scan(&r.RanAt, &r.Tool, &r.OutputKey, &r.Regions, &verdict)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("public issue redaction: %w", err)
	}
	r.Published = verdict == "published"
	v.Redaction = &r
	return nil
}

func (d *DB) attachPublicEvents(ctx context.Context, v *PublicIssueView) error {
	// The actor column is deliberately not selected.
	rows, err := d.pool.Query(ctx, `
		SELECT seq, kind, at, COALESCE(detail,''), hash
		  FROM issue_events WHERE issue_id = $1::uuid ORDER BY seq`, v.IssueID)
	if err != nil {
		return fmt.Errorf("public issue history: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var e PublicEvent
		if err := rows.Scan(&e.Seq, &e.Kind, &e.At, &e.Detail, &e.Hash); err != nil {
			return fmt.Errorf("scan public event: %w", err)
		}
		v.Events = append(v.Events, e)
	}
	return rows.Err()
}

func (d *DB) attachDeadline(ctx context.Context, v *PublicIssueView) error {
	key, ok := deadlineKeys[v.Category]
	if !ok {
		return nil
	}
	var value, citation string
	err := d.pool.QueryRow(ctx, `
		SELECT value::text, citation FROM legal_constants
		 WHERE key = $1 AND effective_from <= now()::date
		   AND (effective_to IS NULL OR effective_to >= now()::date)
		 ORDER BY effective_from DESC LIMIT 1`, key).Scan(&value, &citation)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("public issue deadline %s: %w", key, err)
	}
	hours, convErr := strconv.Atoi(strings.Trim(value, `"`))
	if convErr != nil || citation == "" {
		// A deadline with no citation is an assertion. It is dropped rather
		// than shown, which is the same rule the ward page follows.
		return nil
	}
	v.DeadlineHours, v.DeadlineCitation = hours, citation
	return nil
}

func isBadUUID(err error) bool {
	return strings.Contains(err.Error(), "invalid input syntax for type uuid") ||
		strings.Contains(err.Error(), "invalid UUID")
}
