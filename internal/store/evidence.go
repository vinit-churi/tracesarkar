package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/vinit-churi/tracesarkar/internal/evidence"
)

// EvidencePhoto names a redacted copy to put in the pack, with the record of
// what obscured what.
type EvidencePhoto struct {
	Key       string
	Tool      string
	Regions   int
	Published bool
	TakenAt   *string
	AccuracyM float64
}

// EvidenceFor compiles one issue for the person who reported it.
//
// Account-scoped, and scoped by *sighting* rather than by ownership of the
// issue: an issue belongs to nobody, but the people who photographed it are
// the ones who can act on it. This is the same rule that decides who may
// record a lodging ([D120]).
//
// Somebody with no sighting is told the issue does not exist rather than that
// they may not have it — 404, never 403. Which issues a stranger may not read
// is itself information.
//
// Unlike every other surface this one carries the exact position and the
// complaint reference. That is the point of it: it exists to be attached to a
// complaint, and a complaint needs the place and the number. Everything in it
// is already the reporter's own.
func (d *DB) EvidenceFor(ctx context.Context, issueID, accountID string) (evidence.Record, []EvidencePhoto, bool, error) {
	mine, err := d.ReporterOf(ctx, issueID, accountID)
	if err != nil {
		if isBadUUID(err) {
			return evidence.Record{}, nil, false, nil
		}
		return evidence.Record{}, nil, false, err
	}
	if !mine {
		return evidence.Record{}, nil, false, nil
	}

	var r evidence.Record
	err = d.pool.QueryRow(ctx, `
		SELECT i.id::text, i.category, COALESCE(i.subcategory,''), i.status,
		       COALESCE(i.authority,''), COALESCE(i.ward,''),
		       (SELECT count(*) FROM issue_reports ir WHERE ir.issue_id = i.id),
		       i.first_seen_at, i.last_seen_at, i.notified_at, COALESCE(i.reference,'')
		  FROM issues i WHERE i.id = $1::uuid`, issueID).Scan(
		&r.IssueID, &r.Category, &r.Subcategory, &r.Status,
		&r.Authority, &r.Ward, &r.Sightings,
		&r.FirstSeen, &r.LastSeen, &r.LodgedOn, &r.Reference)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return evidence.Record{}, nil, false, nil
		}
		return evidence.Record{}, nil, false, fmt.Errorf("evidence issue: %w", err)
	}

	// The reporter's own earliest sighting is the one the pack is built from:
	// it carries the position that complaint will quote, and it is theirs.
	var reportID string
	err = d.pool.QueryRow(ctx, `
		SELECT r.id::text, ST_Y(r.location::geometry), ST_X(r.location::geometry),
		       COALESCE(r.location_accuracy_m, 0)::float8
		  FROM issue_reports ir
		  JOIN reports r ON r.id = ir.report_id
		 WHERE ir.issue_id = $1::uuid AND r.account_id::text = $2
		 ORDER BY r.created_at LIMIT 1`, issueID, accountID).
		Scan(&reportID, &r.Lat, &r.Lon, &r.AccuracyM)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return evidence.Record{}, nil, false, fmt.Errorf("evidence report: %w", err)
	}

	if reportID != "" {
		e, err := d.EnrichmentFor(ctx, reportID)
		if err != nil {
			return evidence.Record{}, nil, false, err
		}
		r.WardBasis = e.WardBasis
		// Hard rule 2: the attribution is carried only when it names its
		// source. A match with no source is one the pack states as absent.
		if strings.TrimSpace(e.ContractSource) != "" && e.ContractRetrievedAt != nil {
			r.Contract = &evidence.Contract{
				WorkCode: e.WorkCode, Contractor: e.ContractorName,
				Confidence: e.AttrConfidence, DistanceM: e.DistanceM,
				Basis: e.ContractBasis, SourceRef: d.sourceReference(ctx, e.ContractSource),
				RetrievedAt: *e.ContractRetrievedAt,
			}
		}
	}

	if err := d.attachEvidenceDesk(ctx, &r); err != nil {
		return evidence.Record{}, nil, false, err
	}

	events, err := d.IssueTimeline(ctx, issueID)
	if err != nil {
		return evidence.Record{}, nil, false, err
	}
	// issue.Event carries an Actor, and it is deliberately not copied: this
	// document goes to a desk, and who else photographed the same pothole is
	// not that desk's business. The sequence is the position in the chain.
	for i, e := range events {
		r.Events = append(r.Events, evidence.Event{
			Seq: i + 1, Kind: e.Kind, At: e.At, Detail: e.Detail, Hash: e.Hash,
		})
	}

	photos, err := d.evidencePhotos(ctx, issueID, accountID)
	if err != nil {
		return evidence.Record{}, nil, false, err
	}
	return r, photos, true, nil
}

// sourceReference turns a register key into something a reader can open.
//
// `report_attribution.source_id` is a key into the source register —
// "bmc_roads_api" — which is enough for us and nothing at all for the desk
// this document is handed to. Hard rule 2 asks for a reference a reader can
// check, so the register's own name and URL are resolved here. The key is kept
// alongside, because it is what joins this document back to the register.
func (d *DB) sourceReference(ctx context.Context, sourceID string) string {
	var name, url string
	err := d.pool.QueryRow(ctx,
		`SELECT name, COALESCE(url, '') FROM sources WHERE id = $1`, sourceID).
		Scan(&name, &url)
	if err != nil {
		// A key with no register entry is still the honest answer; inventing a
		// URL for it would be worse than printing the key.
		return sourceID
	}
	switch {
	case url != "" && name != "":
		return name + " (" + sourceID + "), " + url
	case url != "":
		return sourceID + ", " + url
	case name != "":
		return name + " (" + sourceID + ")"
	}
	return sourceID
}

func (d *DB) attachEvidenceDesk(ctx context.Context, r *evidence.Record) error {
	if r.Authority == "" || r.Ward == "" || r.Category == "" {
		return nil
	}
	routing, found, err := d.RoutingFor(ctx, r.Authority, r.Ward, r.Category, r.Subcategory)
	if err != nil {
		return err
	}
	if found && strings.TrimSpace(routing.SourceRef) != "" {
		r.Desk = &evidence.Desk{
			Department: routing.Department, Officer: routing.Officer,
			Address: routing.Office, Phone: routing.OfficePhone,
			Email: routing.OfficeEmail, SourceRef: routing.SourceRef,
		}
		if err := d.pool.QueryRow(ctx, `
			SELECT retrieved_at FROM authority_departments
			 WHERE authority = $1 AND ward = $2 AND category = $3
			 ORDER BY effective_from DESC, created_at LIMIT 1`,
			r.Authority, r.Ward, r.Category).Scan(&r.Desk.RetrievedAt); err != nil &&
			!errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("evidence desk retrieval date: %w", err)
		}
	}
	r.DeadlineHours, r.DeadlineCitation = routing.DeadlineHours, routing.DeadlineCitation
	return nil
}

// evidencePhotos lists this account's own redacted copies for the issue.
//
// Redacted, and only where the gate published them. The pack is a document to
// hand to somebody else, which is the case redaction exists for, and the
// archival original stays break-glass however much more useful it would be.
func (d *DB) evidencePhotos(ctx context.Context, issueID, accountID string) ([]EvidencePhoto, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT COALESCE(m.redaction_output_key,''), COALESCE(m.redaction_tool,''),
		       COALESCE(m.redaction_regions,0), m.redaction_verdict = 'published',
		       to_char(m.exif_taken_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		       COALESCE(r.location_accuracy_m, 0)::float8
		  FROM issue_reports ir
		  JOIN reports r ON r.id = ir.report_id
		  JOIN report_media m ON m.report_id = r.id
		 WHERE ir.issue_id = $1::uuid AND r.account_id::text = $2
		   AND m.redaction_verdict IS NOT NULL
		 ORDER BY m.created_at`, issueID, accountID)
	if err != nil {
		return nil, fmt.Errorf("evidence photographs: %w", err)
	}
	defer rows.Close()

	var out []EvidencePhoto
	for rows.Next() {
		var p EvidencePhoto
		if err := rows.Scan(&p.Key, &p.Tool, &p.Regions, &p.Published,
			&p.TakenAt, &p.AccuracyM); err != nil {
			return nil, fmt.Errorf("scan evidence photograph: %w", err)
		}
		if p.Key == "" || !p.Published {
			continue
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
