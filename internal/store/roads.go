package store

import (
	"context"
	"fmt"
	"time"
)

// NewRoadSegment is one road work's geometry, ready to be projected out of the
// works archive into a queryable table.
type NewRoadSegment struct {
	WorkID       string
	WorkCode     string
	Ward         string
	LocationName string
	// GeoJSON is a MultiLineString in WGS84, as produced by
	// roads.Segment.GeoJSON.
	GeoJSON   string
	StartDate *time.Time
	EndDate   *time.Time
}

// RoadMatch is a road segment near a point, with how near.
type RoadMatch struct {
	WorkID       string     `json:"work_id"`
	WorkCode     string     `json:"work_code"`
	Ward         string     `json:"ward"`
	LocationName string     `json:"location_name"`
	DistanceM    float64    `json:"distance_m"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
}

// SaveRoadSegment stores one segment's geometry.
//
// It is idempotent on the work it came from: road_segments is a projection of
// the works archive and is rebuilt routinely, so a reload must update in place
// rather than accumulate a row per run.
func (d *DB) SaveRoadSegment(ctx context.Context, in NewRoadSegment) error {
	_, err := d.pool.Exec(ctx, `
		INSERT INTO road_segments (id, work_id, work_code, ward, location_name,
		                           geom, start_date, end_date, loaded_at)
		VALUES (gen_random_uuid(), $1::uuid, $2, $3, NULLIF($4,''),
		        ST_GeogFromWKB(ST_AsBinary(ST_GeomFromGeoJSON($5))), $6, $7, now())
		ON CONFLICT (work_id) DO UPDATE SET
		    work_code     = EXCLUDED.work_code,
		    ward          = EXCLUDED.ward,
		    location_name = EXCLUDED.location_name,
		    geom          = EXCLUDED.geom,
		    start_date    = EXCLUDED.start_date,
		    end_date      = EXCLUDED.end_date,
		    loaded_at     = now()`,
		in.WorkID, in.WorkCode, in.Ward, in.LocationName, in.GeoJSON,
		in.StartDate, in.EndDate)
	if err != nil {
		return fmt.Errorf("save road segment %s: %w", in.WorkCode, err)
	}
	return nil
}

// NearestRoadSegments returns the segments within radiusM of a point, nearest
// first.
//
// This is the attribution join. It deliberately returns several candidates
// rather than one answer: at a junction two roads are genuinely close, and
// which one a defect belongs to is a judgement the caller has to make
// explicitly, with the distances in hand, rather than have silently made for
// it by a LIMIT 1.
func (d *DB) NearestRoadSegments(ctx context.Context, lat, lon, radiusM float64, limit int) ([]RoadMatch, error) {
	if limit <= 0 {
		limit = 5
	}
	// ST_DWithin on geography uses the GiST index; ST_Distance then gives
	// metres on the spheroid.
	rows, err := d.pool.Query(ctx, `
		WITH p AS (SELECT ST_SetSRID(ST_MakePoint($2::float8, $1::float8), 4326)::geography AS g)
		SELECT s.work_id::text, s.work_code, s.ward, COALESCE(s.location_name, ''),
		       ST_Distance(s.geom, p.g) AS distance_m,
		       s.start_date, s.end_date
		  FROM road_segments s, p
		 WHERE ST_DWithin(s.geom, p.g, $3::float8)
		 ORDER BY distance_m
		 LIMIT $4`, lat, lon, radiusM, limit)
	if err != nil {
		return nil, fmt.Errorf("nearest road segments: %w", err)
	}
	defer rows.Close()

	var out []RoadMatch
	for rows.Next() {
		var m RoadMatch
		if err := rows.Scan(&m.WorkID, &m.WorkCode, &m.Ward, &m.LocationName,
			&m.DistanceM, &m.StartDate, &m.EndDate); err != nil {
			return nil, fmt.Errorf("scan road match: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("nearest road segments: %w", err)
	}
	return out, nil
}

// CountRoadSegments reports how many segments are loaded, by ward.
func (d *DB) CountRoadSegments(ctx context.Context) (map[string]int, error) {
	rows, err := d.pool.Query(ctx, `SELECT ward, count(*) FROM road_segments GROUP BY ward`)
	if err != nil {
		return nil, fmt.Errorf("count road segments: %w", err)
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var ward string
		var n int
		if err := rows.Scan(&ward, &n); err != nil {
			return nil, fmt.Errorf("scan count: %w", err)
		}
		out[ward] = n
	}
	return out, rows.Err()
}

// WorkRecord is one row of the works archive.
type WorkRecord struct {
	ID      string
	Current map[string]any
}

// WorkRecords returns the works archive for projection into road geometry.
//
// Only rows that still exist upstream are returned: a work BMC has withdrawn
// should stop producing an attribution, and the archive keeps the history
// either way.
func (d *DB) WorkRecords(ctx context.Context) ([]WorkRecord, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT id::text, current FROM works WHERE vanished_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("read works: %w", err)
	}
	defer rows.Close()

	var out []WorkRecord
	for rows.Next() {
		var rec WorkRecord
		if err := rows.Scan(&rec.ID, &rec.Current); err != nil {
			return nil, fmt.Errorf("scan work: %w", err)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}
