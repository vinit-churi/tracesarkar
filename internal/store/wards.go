package store

import (
	"context"
	"fmt"
	"time"
)

// NewWardBoundary is one ward polygon, ready to be stored.
type NewWardBoundary struct {
	Ward        string
	Authority   string
	GeoJSON     string
	SourceID    string
	RetrievedAt *time.Time
	ArchiveKey  string
	Vintage     string
}

// WardHit is a ward a point falls in, or near.
type WardHit struct {
	Ward      string `json:"ward"`
	Authority string `json:"authority"`
	Inside    bool   `json:"inside"`
	// DistanceToEdgeM is the distance to the ward's boundary *line*.
	// ST_Distance to the polygon is zero anywhere inside it, which tells you
	// nothing about whether the answer is safe; nearness to the edge does.
	DistanceToEdgeM float64 `json:"distance_to_edge_m"`
	Vintage         string  `json:"vintage"`
}

// SaveWardBoundary stores one ward. Idempotent: the boundary set is reloaded
// whenever BMC republishes it.
func (d *DB) SaveWardBoundary(ctx context.Context, in NewWardBoundary) error {
	authority := in.Authority
	if authority == "" {
		authority = "BMC"
	}
	_, err := d.pool.Exec(ctx, `
		INSERT INTO ward_boundaries (id, ward, authority, geom, source_id,
		                             retrieved_at, archive_key, vintage, loaded_at)
		VALUES (gen_random_uuid(), $1, $2,
		        ST_GeogFromWKB(ST_AsBinary(ST_Multi(ST_GeomFromGeoJSON($3)))),
		        NULLIF($4,''), $5, NULLIF($6,''), NULLIF($7,''), now())
		ON CONFLICT (authority, ward) DO UPDATE SET
		    geom         = EXCLUDED.geom,
		    source_id    = EXCLUDED.source_id,
		    retrieved_at = EXCLUDED.retrieved_at,
		    archive_key  = EXCLUDED.archive_key,
		    vintage      = EXCLUDED.vintage,
		    loaded_at    = now()`,
		in.Ward, authority, in.GeoJSON, in.SourceID, in.RetrievedAt,
		in.ArchiveKey, in.Vintage)
	if err != nil {
		return fmt.Errorf("save ward boundary %s: %w", in.Ward, err)
	}
	return nil
}

// WardsNear returns the ward containing the point, and any ward whose boundary
// runs within toleranceM of it.
//
// Several results is the interesting case: a point near a ward line is exactly
// where routing goes wrong, and the caller needs to know that rather than be
// handed the first answer.
func (d *DB) WardsNear(ctx context.Context, lat, lon, toleranceM float64) ([]WardHit, error) {
	rows, err := d.pool.Query(ctx, `
		WITH p AS (SELECT ST_SetSRID(ST_MakePoint($2::float8, $1::float8), 4326)::geography AS g)
		SELECT w.ward, w.authority,
		       ST_Intersects(w.geom, p.g) AS inside,
		       ST_Distance(ST_Boundary(w.geom::geometry)::geography, p.g) AS edge_m,
		       COALESCE(w.vintage, '')
		  FROM ward_boundaries w, p
		 WHERE ST_DWithin(w.geom, p.g, $3::float8)
		 ORDER BY inside DESC, edge_m`, lat, lon, toleranceM)
	if err != nil {
		return nil, fmt.Errorf("wards near point: %w", err)
	}
	defer rows.Close()

	var out []WardHit
	for rows.Next() {
		var h WardHit
		if err := rows.Scan(&h.Ward, &h.Authority, &h.Inside, &h.DistanceToEdgeM, &h.Vintage); err != nil {
			return nil, fmt.Errorf("scan ward hit: %w", err)
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// CountWardBoundaries reports how many wards are loaded.
func (d *DB) CountWardBoundaries(ctx context.Context) (int, error) {
	var n int
	err := d.pool.QueryRow(ctx, `SELECT count(*) FROM ward_boundaries`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count ward boundaries: %w", err)
	}
	return n, nil
}
