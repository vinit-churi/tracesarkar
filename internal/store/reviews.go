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
	AgreesWithExpectation *bool `json:"agrees_with_expectation,omitempty"`
	// RoadGeoJSON is the matched road, so the reviewer sees the shape the join
	// actually chose rather than a pin and a name.
	RoadGeoJSON string `json:"road_geojson"`
	// Priority says why this item is worth looking at. A queue sorted by it
	// puts the informative cases first, so an hour of attention is not spent
	// confirming answers the machine could already check itself.
	Priority string `json:"priority"`
	// RivalLocation is the next road that is nearly as close, when there is
	// one. It is the actual question in an ambiguous case.
	RivalLocation string   `json:"rival_location,omitempty"`
	RivalDistance *float64 `json:"rival_distance_m,omitempty"`
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
	// The second-nearest road that would give a *different* answer. When it is
	// close, that is the whole question; when there is none, the item is
	// routine and the machine already knows the answer.
	rows, err := d.pool.Query(ctx, `
		WITH item AS (
		  SELECT r.*, ST_SetSRID(ST_MakePoint(r.lon, r.lat), 4326)::geography AS g
		    FROM attribution_reviews r WHERE r.verdict IS NULL
		), rival AS (
		  SELECT i.id,
		         s.location_name AS rival_name,
		         ST_Distance(s.geom, i.g) AS rival_distance
		    FROM item i
		    JOIN LATERAL (
		      SELECT s2.location_name, s2.geom
		        FROM road_segments s2
		       WHERE s2.work_id IS DISTINCT FROM i.matched_work_id
		         AND (s2.work_code, COALESCE(s2.contractor_name,'')) IS DISTINCT FROM
		             (i.matched_work_code, COALESCE(i.matched_contractor,''))
		         AND ST_DWithin(s2.geom, i.g, 60)
		       ORDER BY ST_Distance(s2.geom, i.g)
		       LIMIT 1
		    ) s ON true
		)
		SELECT i.id::text, i.kind, i.lat, i.lon, i.accuracy_m::float8, i.confidence,
		       COALESCE(i.matched_work_code,''), COALESCE(i.matched_contractor,''),
		       COALESCE(i.matched_location,''), i.distance_m, i.basis,
		       CASE WHEN i.kind = 'probe' AND i.expected_work_id IS NOT NULL
		            THEN (i.matched_work_id IS NOT DISTINCT FROM i.expected_work_id)
		       END AS agrees,
		       COALESCE(ST_AsGeoJSON(s.geom::geometry), '') AS road,
		       CASE
		         WHEN i.kind = 'probe' AND i.expected_work_id IS NOT NULL
		              AND i.matched_work_id IS DISTINCT FROM i.expected_work_id THEN $2
		         WHEN v.rival_distance IS NOT NULL
		              AND v.rival_distance - COALESCE(i.distance_m, 0) < i.accuracy_m THEN $3
		         WHEN i.matched_work_id IS NULL OR i.distance_m > 15 THEN $4
		         ELSE $5
		       END AS priority,
		       COALESCE(v.rival_name, ''), v.rival_distance
		  FROM item i
		  LEFT JOIN road_segments s ON s.work_id = i.matched_work_id
		  LEFT JOIN rival v ON v.id = i.id
		 ORDER BY CASE
		     WHEN i.kind = 'probe' AND i.expected_work_id IS NOT NULL
		          AND i.matched_work_id IS DISTINCT FROM i.expected_work_id THEN 0
		     WHEN v.rival_distance IS NOT NULL
		          AND v.rival_distance - COALESCE(i.distance_m, 0) < i.accuracy_m THEN 1
		     WHEN i.matched_work_id IS NULL OR i.distance_m > 15 THEN 2
		     ELSE 3
		   END, i.created_at
		 LIMIT $1`, limit, PriorityDisagrees, PriorityAmbiguous, PriorityDistant, PriorityRoutine)
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
			&it.AgreesWithExpectation, &it.RoadGeoJSON,
			&it.Priority, &it.RivalLocation, &it.RivalDistance); err != nil {
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


// ExpectationAgreement reports whether the join returned the segment a probe
// was generated from. Null for a real report, which has no expected answer —
// only a person can say what road is in a photograph.
func (d *DB) ExpectationAgreement(ctx context.Context, id string) (*bool, error) {
	var agreed *bool
	err := d.pool.QueryRow(ctx, `
		SELECT CASE WHEN kind = 'probe' AND expected_work_id IS NOT NULL
		            THEN (matched_work_id IS NOT DISTINCT FROM expected_work_id)
		       END
		  FROM attribution_reviews WHERE id = $1::uuid`, id).Scan(&agreed)
	if err != nil {
		return nil, fmt.Errorf("expectation agreement: %w", err)
	}
	return agreed, nil
}

// ReviewPriority explains why an item is worth a person's attention.
// Surfaced so the page can say what it is asking and why this one.
const (
	PriorityDisagrees = "disagrees" // the join chose a different road from the probe's own
	PriorityAmbiguous = "ambiguous" // a second road is nearly as close
	PriorityDistant   = "distant"   // the point is far from anything, or matched nothing
	PriorityRoutine   = "routine"   // one obvious road; little to learn
)
