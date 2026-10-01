package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/vinit-churi/tracesarkar/internal/attribute"
	"github.com/vinit-churi/tracesarkar/internal/jurisdiction"
)

// ReportPoint is a capture's position, which is what the two spatial systems
// need and all they need.
type ReportPoint struct {
	ReportID  string
	Lat, Lon  float64
	AccuracyM float64
}

// UnenrichedReports returns captures missing a jurisdiction or an attribution.
//
// A report with a ward and no contract is finished, not partial: about half of
// Borivali has no contract data, and saying so is a real answer. So the two
// are tracked separately and a report is only pending for what it is missing.
func (d *DB) UnenrichedReports(ctx context.Context, limit int) ([]ReportPoint, error) {
	if limit <= 0 || limit > 200 {
		limit = 25
	}
	rows, err := d.pool.Query(ctx, `
		SELECT r.id::text,
		       ST_Y(r.location::geometry), ST_X(r.location::geometry),
		       COALESCE(r.location_accuracy_m, 0)::float8
		  FROM reports r
		 WHERE NOT EXISTS (SELECT 1 FROM report_jurisdiction j
		                    WHERE j.report_id = r.id AND j.error IS NULL)
		    OR NOT EXISTS (SELECT 1 FROM report_attribution a
		                    WHERE a.report_id = r.id AND a.error IS NULL)
		 ORDER BY r.created_at
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("find unenriched reports: %w", err)
	}
	defer rows.Close()

	var out []ReportPoint
	for rows.Next() {
		var p ReportPoint
		if err := rows.Scan(&p.ReportID, &p.Lat, &p.Lon, &p.AccuracyM); err != nil {
			return nil, fmt.Errorf("scan report point: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SaveJurisdiction records which authority owns a capture.
func (d *DB) SaveJurisdiction(ctx context.Context, reportID string, r jurisdiction.Resolution, vintage, errText string) error {
	alts, _ := json.Marshal(r.Alternatives)
	_, err := d.pool.Exec(ctx, `
		INSERT INTO report_jurisdiction (id, report_id, ward, authority, confidence,
		    needs_question, basis, alternatives, boundary_vintage, error)
		VALUES (gen_random_uuid(), $1::uuid, NULLIF($2,''), NULLIF($3,''), $4,
		        $5, $6, $7::jsonb, NULLIF($8,''), NULLIF($9,''))`,
		reportID, r.Ward, r.Authority, string(r.Confidence), r.NeedsQuestion,
		r.Basis, string(alts), vintage, errText)
	if err != nil {
		return fmt.Errorf("save jurisdiction for %s: %w", reportID, err)
	}
	return nil
}

// SaveAttribution records which contract covers a capture.
func (d *DB) SaveAttribution(ctx context.Context, reportID string, a attribute.Attribution, source string, retrieved *time.Time, errText string) error {
	alts, _ := json.Marshal(a.Alternatives)

	var workID *string
	var workCode, contractor, location string
	var distance *float64
	if a.Match != nil {
		id := a.Match.WorkID
		workID = &id
		workCode = a.Match.WorkCode
		contractor = a.Match.ContractorName
		location = a.Match.LocationName
		dm := a.Match.DistanceM
		distance = &dm
	}

	_, err := d.pool.Exec(ctx, `
		INSERT INTO report_attribution (id, report_id, confidence, matched_work_id,
		    work_code, contractor_name, location_name, distance_m, search_radius_m,
		    basis, would_change, alternatives, source_id, retrieved_at, error)
		VALUES (gen_random_uuid(), $1::uuid, $2, $3::uuid, NULLIF($4,''),
		        NULLIF($5,''), NULLIF($6,''), $7, $8, $9, NULLIF($10,''),
		        $11::jsonb, NULLIF($12,''), $13, NULLIF($14,''))`,
		reportID, string(a.Confidence), workID, workCode, contractor, location,
		distance, a.SearchRadiusM, a.Basis, a.WouldChange, string(alts),
		source, retrieved, errText)
	if err != nil {
		return fmt.Errorf("save attribution for %s: %w", reportID, err)
	}
	return nil
}

// ReportEnrichment is everything the platform has concluded about a capture.
type ReportEnrichment struct {
	Ward             string  `json:"ward,omitempty"`
	Authority        string  `json:"authority,omitempty"`
	WardConfidence   string  `json:"ward_confidence,omitempty"`
	WardBasis        string  `json:"ward_basis,omitempty"`
	NeedsQuestion    bool    `json:"needs_question"`
	ContractorName   string  `json:"contractor_name,omitempty"`
	WorkCode         string  `json:"work_code,omitempty"`
	LocationName     string  `json:"location_name,omitempty"`
	DistanceM        float64 `json:"distance_m,omitempty"`
	ContractBasis    string  `json:"contract_basis,omitempty"`
	ContractSource   string  `json:"contract_source,omitempty"`
	AttrConfidence   string  `json:"attribution_confidence,omitempty"`
}

// EnrichmentFor returns the latest conclusions about a capture. Missing parts
// stay empty rather than erroring: a report with a ward and no contract is a
// complete answer.
func (d *DB) EnrichmentFor(ctx context.Context, reportID string) (ReportEnrichment, error) {
	var e ReportEnrichment

	err := d.pool.QueryRow(ctx, `
		SELECT COALESCE(ward,''), COALESCE(authority,''), confidence,
		       COALESCE(basis,''), needs_question
		  FROM report_jurisdiction
		 WHERE report_id = $1::uuid AND error IS NULL
		 ORDER BY created_at DESC LIMIT 1`, reportID).Scan(
		&e.Ward, &e.Authority, &e.WardConfidence, &e.WardBasis, &e.NeedsQuestion)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return e, fmt.Errorf("read jurisdiction: %w", err)
	}

	err = d.pool.QueryRow(ctx, `
		SELECT confidence, COALESCE(contractor_name,''), COALESCE(work_code,''),
		       COALESCE(location_name,''), COALESCE(distance_m,0),
		       COALESCE(basis,''), COALESCE(source_id,'')
		  FROM report_attribution
		 WHERE report_id = $1::uuid AND error IS NULL
		 ORDER BY created_at DESC LIMIT 1`, reportID).Scan(
		&e.AttrConfidence, &e.ContractorName, &e.WorkCode, &e.LocationName,
		&e.DistanceM, &e.ContractBasis, &e.ContractSource)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return e, fmt.Errorf("read attribution: %w", err)
	}

	return e, nil
}

// ReportDetail reads one capture and everything concluded about it.
//
// Missing enrichment is not an error: a capture whose classification has not
// run yet is still a capture, and the photograph is safe either way.
func (d *DB) ReportDetail(ctx context.Context, id string) (ReportDetail, bool, error) {
	var out ReportDetail
	err := d.pool.QueryRow(ctx, `
		SELECT r.id::text, r.account_id::text, r.status::text,
		       to_char(r.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		       ST_Y(r.location::geometry), ST_X(r.location::geometry),
		       COALESCE(r.location_accuracy_m, 0)::float8
		  FROM reports r WHERE r.id = $1::uuid`, id).Scan(
		&out.ID, &out.AccountID, &out.Status, &out.CreatedAt,
		&out.Lat, &out.Lon, &out.AccuracyM)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReportDetail{}, false, nil
		}
		return ReportDetail{}, false, fmt.Errorf("read report: %w", err)
	}

	if dec, ok, err := d.LatestClassification(ctx, id); err != nil {
		return out, true, err
	} else if ok {
		out.Classification = &dec
	}

	e, err := d.EnrichmentFor(ctx, id)
	if err != nil {
		return out, true, err
	}
	out.Ward, out.Authority = e.Ward, e.Authority
	out.WardConfidence, out.WardBasis = e.WardConfidence, e.WardBasis
	out.NeedsQuestion = e.NeedsQuestion
	out.ContractorName, out.WorkCode = e.ContractorName, e.WorkCode
	out.RoadName, out.DistanceM = e.LocationName, e.DistanceM
	out.ContractBasis, out.ContractSource = e.ContractBasis, e.ContractSource
	out.AttrConfidence = e.AttrConfidence

	return out, true, nil
}
