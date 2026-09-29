package store

import (
	"context"
	"fmt"

	"github.com/vinit-churi/tracesarkar/internal/attribute"
)

// NewReviewItem is one point, and the answer the join gave for it, queued for
// a person to judge.
type NewReviewItem struct {
	Kind              string // "probe" or "report"
	ReportID          *string
	Lat, Lon          float64
	AccuracyM         float64
	Confidence        string
	MatchedWorkID     *string
	MatchedWorkCode   string
	MatchedContractor string
	MatchedLocation   string
	DistanceM         *float64
	Basis             string
	ExpectedWorkID    *string
}

// ReviewItem is a queued answer, as shown to the reviewer.
type ReviewItem struct {
	ID                string   `json:"id"`
	Kind              string   `json:"kind"`
	Lat               float64  `json:"lat"`
	Lon               float64  `json:"lon"`
	AccuracyM         float64  `json:"accuracy_m"`
	Confidence        string   `json:"confidence"`
	MatchedWorkCode   string   `json:"matched_work_code"`
	MatchedContractor string   `json:"matched_contractor"`
	MatchedLocation   string   `json:"matched_location"`
	DistanceM         *float64 `json:"distance_m"`
	Basis             string   `json:"basis"`
	// AgreesWithExpectation is set for probes: whether the join returned the
	// segment the point was generated from. It is a hint for the reviewer,
	// never a substitute for their judgement — the expectation can be wrong.
	AgreesWithExpectation *bool `json:"agrees_with_expectation"`
	// RoadGeoJSON is the matched road, so the reviewer sees the shape the join
	// actually chose rather than a pin and a name.
	RoadGeoJSON string `json:"road_geojson"`
}

// SaveReviewItem queues one answer.
func (d *DB) SaveReviewItem(ctx context.Context, in NewReviewItem) error {
	_, err := d.pool.Exec(ctx, `
		INSERT INTO attribution_reviews (id, kind, report_id, lat, lon, accuracy_m,
		    confidence, matched_work_id, matched_work_code, matched_contractor,
		    matched_location, distance_m, basis, expected_work_id)
		VALUES (gen_random_uuid(), $1, $2::uuid, $3, $4, $5::numeric, $6, $7::uuid,
		        NULLIF($8,''), NULLIF($9,''), NULLIF($10,''), $11, $12, $13::uuid)`,
		in.Kind, in.ReportID, in.Lat, in.Lon, in.AccuracyM, in.Confidence,
		in.MatchedWorkID, in.MatchedWorkCode, in.MatchedContractor,
		in.MatchedLocation, in.DistanceM, in.Basis, in.ExpectedWorkID)
	if err != nil {
		return fmt.Errorf("queue review item: %w", err)
	}
	return nil
}

// PendingReviewItems returns answers nobody has judged yet.
func (d *DB) PendingReviewItems(ctx context.Context, limit int) ([]ReviewItem, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := d.pool.Query(ctx, `
		SELECT r.id::text, r.kind, r.lat, r.lon, r.accuracy_m::float8, r.confidence,
		       COALESCE(r.matched_work_code,''), COALESCE(r.matched_contractor,''),
		       COALESCE(r.matched_location,''), r.distance_m, r.basis,
		       CASE WHEN r.kind = 'probe' AND r.expected_work_id IS NOT NULL
		            THEN (r.matched_work_id IS NOT DISTINCT FROM r.expected_work_id)
		       END AS agrees,
		       COALESCE(ST_AsGeoJSON(s.geom::geometry), '') AS road
		  FROM attribution_reviews r
		  LEFT JOIN road_segments s ON s.work_id = r.matched_work_id
		 WHERE r.verdict IS NULL
		 ORDER BY r.created_at
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("pending review items: %w", err)
	}
	defer rows.Close()

	var out []ReviewItem
	for rows.Next() {
		var it ReviewItem
		if err := rows.Scan(&it.ID, &it.Kind, &it.Lat, &it.Lon, &it.AccuracyM,
			&it.Confidence, &it.MatchedWorkCode, &it.MatchedContractor,
			&it.MatchedLocation, &it.DistanceM, &it.Basis,
			&it.AgreesWithExpectation, &it.RoadGeoJSON); err != nil {
			return nil, fmt.Errorf("scan review item: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// RecordVerdict stores what a person decided.
func (d *DB) RecordVerdict(ctx context.Context, id, verdict, note, reviewerID string) error {
	switch verdict {
	case "correct", "wrong", "unsure":
	default:
		return fmt.Errorf("verdict %q is not one of correct, wrong, unsure", verdict)
	}
	tag, err := d.pool.Exec(ctx, `
		UPDATE attribution_reviews
		   SET verdict = $2, verdict_note = NULLIF($3,''),
		       reviewed_at = now(), reviewed_by = NULLIF($4,'')::uuid
		 WHERE id = $1::uuid`, id, verdict, note, reviewerID)
	if err != nil {
		return fmt.Errorf("record verdict: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no review item %s", id)
	}
	return nil
}

// ReviewSummary is how the queue stands.
func (d *DB) ReviewSummary(ctx context.Context) (attribute.Summary, error) {
	var s attribute.Summary
	err := d.pool.QueryRow(ctx, `
		SELECT count(*),
		       count(*) FILTER (WHERE verdict = 'correct'),
		       count(*) FILTER (WHERE verdict = 'wrong'),
		       count(*) FILTER (WHERE verdict = 'unsure'),
		       count(*) FILTER (WHERE verdict IS NULL)
		  FROM attribution_reviews`).Scan(
		&s.Total, &s.Correct, &s.Wrong, &s.Unsure, &s.Pending)
	if err != nil {
		return s, fmt.Errorf("review summary: %w", err)
	}
	return s, nil
}

// ProbePoint is a generated point near known geometry, with the segment it
// came from.
type ProbePoint struct {
	Lat, Lon       float64
	ExpectedWorkID string
	OffsetM        float64
}

// GenerateProbes makes points a few metres off the centreline of real road
// geometry, to imitate a defect at the road edge seen through a phone's GPS.
//
// Centreline points would flatter the join: a real capture is never on the
// centreline, and the interesting failures happen at the edges, where the next
// road is closer.
func (d *DB) GenerateProbes(ctx context.Context, ward string, count int, minOffsetM, maxOffsetM float64) ([]ProbePoint, error) {
	rows, err := d.pool.Query(ctx, `
		WITH picked AS (
		  SELECT work_id, geom FROM road_segments
		   WHERE ward = $1 ORDER BY random() LIMIT $2
		), placed AS (
		  SELECT work_id,
		         ST_LineInterpolatePoint(ST_GeometryN(geom::geometry, 1), random())::geography AS on_line,
		         $3::float8 + random() * ($4::float8 - $3::float8) AS offset_m,
		         random() * 2 * pi() AS azimuth
		    FROM picked
		)
		SELECT work_id::text,
		       ST_Y(ST_Project(on_line, offset_m, azimuth)::geometry) AS lat,
		       ST_X(ST_Project(on_line, offset_m, azimuth)::geometry) AS lon,
		       offset_m
		  FROM placed`, ward, count, minOffsetM, maxOffsetM)
	if err != nil {
		return nil, fmt.Errorf("generate probes: %w", err)
	}
	defer rows.Close()

	var out []ProbePoint
	for rows.Next() {
		var p ProbePoint
		if err := rows.Scan(&p.ExpectedWorkID, &p.Lat, &p.Lon, &p.OffsetM); err != nil {
			return nil, fmt.Errorf("scan probe: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

