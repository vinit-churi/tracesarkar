package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/vinit-churi/tracesarkar/internal/issue"
)

// IssueNear is an existing issue close to a point, for the merge decision.
type IssueNear struct {
	IssueID   string
	Category  string
	Status    string
	DistanceM float64
}

// IssuesNear finds issues within a radius of a point, nearest first.
func (d *DB) IssuesNear(ctx context.Context, lat, lon, radiusM float64) ([]IssueNear, error) {
	rows, err := d.pool.Query(ctx, `
		WITH p AS (SELECT ST_SetSRID(ST_MakePoint($2::float8, $1::float8), 4326)::geography g)
		SELECT i.id::text, i.category, i.status, ST_Distance(i.centroid, p.g)
		  FROM issues i, p
		 WHERE ST_DWithin(i.centroid, p.g, $3::float8)
		 ORDER BY ST_Distance(i.centroid, p.g), i.id
		 LIMIT 20`, lat, lon, radiusM)
	if err != nil {
		return nil, fmt.Errorf("issues near: %w", err)
	}
	defer rows.Close()

	var out []IssueNear
	for rows.Next() {
		var n IssueNear
		if err := rows.Scan(&n.IssueID, &n.Category, &n.Status, &n.DistanceM); err != nil {
			return nil, fmt.Errorf("scan issue: %w", err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// AttachReport records a report as a sighting, creating the issue or joining
// an existing one according to the decision already made.
//
// One transaction, because an issue without its first sighting and a sighting
// without its issue are both corrupt states, and the history event has to land
// with whichever happened.
func (d *DB) AttachReport(ctx context.Context, in Attachment) (string, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("attach report: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	issueID := in.Decision.IssueID
	switch in.Decision.Action {
	case issue.Create:
		// The identifier comes from the database rather than a dependency:
		// every other table here is generated the same way.
		if err := tx.QueryRow(ctx, `
			INSERT INTO issues (id, authority, ward, category, subcategory,
			                    centroid, status, first_seen_at, last_seen_at)
			VALUES (gen_random_uuid(), NULLIF($1,''), NULLIF($2,''), $3, NULLIF($4,''),
			        ST_SetSRID(ST_MakePoint($6::float8, $5::float8), 4326),
			        'open', $7, $7)
			RETURNING id::text`,
			in.Authority, in.Ward, in.Category, in.Subcategory,
			in.Lat, in.Lon, in.SeenAt).Scan(&issueID); err != nil {
			return "", fmt.Errorf("create issue: %w", err)
		}

	default:
		// The centroid is the mean of its sightings, so each one sharpens the
		// estimate rather than the first one fixing it forever.
		status := "open"
		if in.Decision.Action == issue.Reopen {
			status = "reopened"
		}
		if _, err := tx.Exec(ctx, `
			UPDATE issues SET
			  last_seen_at = GREATEST(last_seen_at, $2),
			  status = $3,
			  centroid = (
			    SELECT ST_Centroid(ST_Collect(ARRAY[
			      i.centroid::geometry,
			      ST_SetSRID(ST_MakePoint($5::float8, $4::float8), 4326)
			    ]))::geography
			    FROM issues i WHERE i.id = $1::uuid)
			 WHERE id = $1::uuid`,
			issueID, in.SeenAt, status, in.Lat, in.Lon); err != nil {
			return "", fmt.Errorf("join issue: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO issue_reports (issue_id, report_id, distance_m, basis)
		VALUES ($1::uuid, $2::uuid, NULLIF($3,0), $4)
		ON CONFLICT (report_id) DO NOTHING`,
		issueID, in.ReportID, in.Decision.DistanceM, in.Decision.Basis); err != nil {
		return "", fmt.Errorf("record sighting: %w", err)
	}

	kind := "reported"
	if in.Decision.Action == issue.Join {
		kind = "corroborated"
	} else if in.Decision.Action == issue.Reopen {
		kind = "reopened"
	}
	if err := appendEventTx(ctx, tx, issueID, issue.Event{
		Kind: kind, At: in.SeenAt, Detail: in.Decision.Basis, Actor: in.Actor,
	}); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("attach report: %w", err)
	}
	return issueID, nil
}

// Attachment is everything needed to record one sighting.
type Attachment struct {
	ReportID    string
	Authority   string
	Ward        string
	Category    string
	Subcategory string
	Lat, Lon    float64
	SeenAt      time.Time
	Actor       string
	Decision    issue.Decision
}

// appendEventTx writes one event onto the end of an issue's chain.
func appendEventTx(ctx context.Context, tx pgx.Tx, issueID string, e issue.Event) error {
	var seq int
	var prev string
	err := tx.QueryRow(ctx, `
		SELECT seq, hash FROM issue_events
		 WHERE issue_id = $1::uuid ORDER BY seq DESC LIMIT 1`, issueID).
		Scan(&seq, &prev)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("read chain head: %w", err)
	}

	chained := issue.Append([]issue.Event{{Hash: prev}}, e)
	next := chained[len(chained)-1]
	if prev == "" {
		// Nothing before it: the anchor, with no previous hash at all.
		next = issue.Chain([]issue.Event{e})[0]
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO issue_events (id, issue_id, seq, kind, at, detail, actor, prev_hash, hash)
		VALUES (gen_random_uuid(), $1::uuid, $2, $3, $4, NULLIF($5,''), NULLIF($6,''),
		        NULLIF($7,''), $8)`,
		issueID, seq+1, next.Kind, next.At, next.Detail, next.Actor,
		next.PrevHash, next.Hash); err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	return nil
}

// IssueTimeline reads an issue's history, oldest first, for verification.
func (d *DB) IssueTimeline(ctx context.Context, issueID string) ([]issue.Event, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT kind, at, COALESCE(detail,''), COALESCE(actor,''),
		       COALESCE(prev_hash,''), hash
		  FROM issue_events WHERE issue_id = $1::uuid ORDER BY seq`, issueID)
	if err != nil {
		return nil, fmt.Errorf("issue timeline: %w", err)
	}
	defer rows.Close()

	var out []issue.Event
	for rows.Next() {
		var e issue.Event
		if err := rows.Scan(&e.Kind, &e.At, &e.Detail, &e.Actor, &e.PrevHash, &e.Hash); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Unattached is a classified capture that is not yet a sighting of anything.
type Unattached struct {
	ReportID    string
	Authority   string
	Ward        string
	Category    string
	Subcategory string
	Lat, Lon    float64
	AccuracyM   float64
	SeenAt      time.Time
	AccountID   string
}

// UnattachedReports returns captures ready to become or join an issue.
//
// Only classified ones: the category decides which issues a sighting could
// belong to, and merging on position alone would put rubbish and a pothole in
// the same ticket. Only ones whose outcome routes, because a photograph the
// platform could not read is not evidence of a problem anywhere.
func (d *DB) UnattachedReports(ctx context.Context, limit int) ([]Unattached, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := d.pool.Query(ctx, `
		SELECT r.id::text, COALESCE(j.authority,''), COALESCE(j.ward,''),
		       c.category, COALESCE(c.subcategory,''),
		       ST_Y(r.location::geometry), ST_X(r.location::geometry),
		       COALESCE(r.location_accuracy_m, 0)::float8,
		       r.captured_at, COALESCE(r.account_id::text,'')
		  FROM reports r
		  JOIN LATERAL (
		    SELECT category, subcategory, outcome FROM classifications
		     WHERE report_id = r.id AND error IS NULL
		     ORDER BY created_at DESC LIMIT 1
		  ) c ON true
		  LEFT JOIN report_jurisdiction j ON j.report_id = r.id
		 WHERE NOT EXISTS (SELECT 1 FROM issue_reports ir WHERE ir.report_id = r.id)
		   AND c.category <> ''
		   -- A capture the platform could not read, or decided is not a civic
		   -- problem, is not evidence of a problem anywhere.
		   AND c.outcome IN ('accepted', 'not_yet_covered')
		 ORDER BY r.captured_at
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("unattached reports: %w", err)
	}
	defer rows.Close()

	var out []Unattached
	for rows.Next() {
		var u Unattached
		if err := rows.Scan(&u.ReportID, &u.Authority, &u.Ward, &u.Category,
			&u.Subcategory, &u.Lat, &u.Lon, &u.AccuracyM, &u.SeenAt,
			&u.AccountID); err != nil {
			return nil, fmt.Errorf("scan unattached: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// IssueView is an issue as the person who reported it sees it.
type IssueView struct {
	IssueID    string
	Category   string
	Status     string
	Sightings  int
	NotifiedAt *time.Time
	Reference  string
	FirstSeen  time.Time
	LastSeen   time.Time
}

// IssueForReport returns the issue a capture is a sighting of, if it has
// become one yet.
func (d *DB) IssueForReport(ctx context.Context, reportID string) (IssueView, bool, error) {
	var v IssueView
	err := d.pool.QueryRow(ctx, `
		SELECT i.id::text, i.category, i.status,
		       (SELECT count(*) FROM issue_reports r WHERE r.issue_id = i.id),
		       i.notified_at, COALESCE(i.reference,''), i.first_seen_at, i.last_seen_at
		  FROM issues i
		  JOIN issue_reports ir ON ir.issue_id = i.id
		 WHERE ir.report_id = $1::uuid`, reportID).
		Scan(&v.IssueID, &v.Category, &v.Status, &v.Sightings,
			&v.NotifiedAt, &v.Reference, &v.FirstSeen, &v.LastSeen)
	if errors.Is(err, pgx.ErrNoRows) {
		return IssueView{}, false, nil
	}
	if err != nil {
		return IssueView{}, false, fmt.Errorf("issue for report: %w", err)
	}
	return v, true, nil
}

// ErrAlreadyNotified is returned when an issue has already been lodged.
var ErrAlreadyNotified = errors.New("already notified")

// RecordNotified marks that the authority has been told, which is the only
// thing that starts a clock.
//
// The first notification wins. A second one is refused rather than overwriting
// the first, because moving the start would move the deadline — a platform
// whose clock can be restarted by re-submitting is one whose deadlines mean
// nothing, and the earlier lodging is the one the obligation runs from.
func (d *DB) RecordNotified(ctx context.Context, issueID, reference, actor string, at time.Time) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("record notified: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var existing *time.Time
	if err := tx.QueryRow(ctx,
		`SELECT notified_at FROM issues WHERE id = $1::uuid FOR UPDATE`, issueID).
		Scan(&existing); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("read issue: %w", err)
	}
	if existing != nil {
		return ErrAlreadyNotified
	}

	if _, err := tx.Exec(ctx,
		`UPDATE issues SET notified_at = $2, reference = NULLIF($3,'') WHERE id = $1::uuid`,
		issueID, at, reference); err != nil {
		return fmt.Errorf("set notified: %w", err)
	}

	detail := "complaint lodged"
	if reference != "" {
		detail = "complaint lodged, reference " + reference
	}
	if err := appendEventTx(ctx, tx, issueID, issue.Event{
		Kind: "notified", At: at, Detail: detail, Actor: actor,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("record notified: %w", err)
	}
	return nil
}

// ReporterOf reports whether an account has a sighting in an issue.
//
// Only somebody who photographed it may say it has been lodged. The reference
// dates a legal clock, so it is not something a passer-by gets to assert about
// another person's report.
func (d *DB) ReporterOf(ctx context.Context, issueID, accountID string) (bool, error) {
	var n int
	err := d.pool.QueryRow(ctx, `
		SELECT count(*) FROM issue_reports ir
		  JOIN reports r ON r.id = ir.report_id
		 WHERE ir.issue_id = $1::uuid AND r.account_id::text = $2`,
		issueID, accountID).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("reporter of: %w", err)
	}
	return n > 0, nil
}
