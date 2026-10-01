package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

// A point 4 m from a road must match it; a point on the next street must not.
// This is the whole attribution claim in one test.
func TestLiveRoadSegmentsAreFoundByDistance(t *testing.T) {
	ctx, pool, db := liveDB(t)

	// Borrow a real collected work rather than building the whole
	// source -> document -> work chain: road_segments is a projection of that
	// table, so a real row is the honest fixture.
	var workID string
	if err := pool.QueryRow(ctx, `SELECT id FROM works LIMIT 1`).Scan(&workID); err != nil {
		t.Skipf("no collected works to project from: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM road_segments WHERE work_code = 'W-415-test'`)
	})

	// A real Borivali stretch, from BMC's own geometry.
	end := time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC)
	seg := store.NewRoadSegment{
		WorkID:       workID,
		WorkCode:     "W-415-test",
		Ward:         "R/C",
		LocationName: "Derasar to Dead",
		GeoJSON: `{"type":"MultiLineString","coordinates":[[` +
			`[72.8441898127395,19.234093217066558],` +
			`[72.84440287106771,19.233907857530042],` +
			`[72.84478303413812,19.233418822942667]]]}`,
		EndDate: &end,
	}
	if err := db.SaveRoadSegment(ctx, seg); err != nil {
		t.Fatalf("SaveRoadSegment: %v", err)
	}

	// Reloading the same work must not create a second row: the table is a
	// projection of works and is rebuilt routinely.
	if err := db.SaveRoadSegment(ctx, seg); err != nil {
		t.Fatalf("second save: %v", err)
	}
	var rows int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM road_segments WHERE work_id = $1`, workID).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("a reload must not duplicate: got %d rows", rows)
	}

	// A point essentially on the line.
	//
	// The fixture is not assumed to be the single nearest row. Its geometry was
	// copied from BMC's own "Derasar to Dead" stretch, and `ingest roads` has
	// since loaded that very segment, so the probe now sits on two coincident
	// rows at the same distance. Which of them sorts first says nothing about
	// whether the search works — asserting on it tested the tie-break, and
	// started failing the moment real geometry was loaded.
	near, err := db.NearestRoadSegments(ctx, 19.23391, 72.84440, 25, 5)
	if err != nil {
		t.Fatalf("NearestRoadSegments: %v", err)
	}
	if len(near) == 0 {
		t.Fatal("a point on the road must match it")
	}

	var found *store.RoadMatch
	for i := range near {
		if near[i].WorkCode == "W-415-test" {
			found = &near[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("a point on the road must match it; got %d other matches", len(near))
	}
	if found.DistanceM > 25 {
		t.Errorf("distance should be within the buffer: %v m", found.DistanceM)
	}

	// Ordering is the contract the old assertion was standing in for, and the
	// attribution gate depends on it: it reads candidates[0] as the nearest and
	// measures every rival against that one. A wider search is used here
	// deliberately — within 25 m the only matches are coincident, so they
	// cannot demonstrate an order at all.
	spread, err := db.NearestRoadSegments(ctx, 19.23391, 72.84440, 250, 5)
	if err != nil {
		t.Fatalf("NearestRoadSegments (wide): %v", err)
	}
	if len(spread) < 3 {
		t.Fatalf("expected several roads within 250 m to order; got %d", len(spread))
	}
	var distinct int
	for i := 1; i < len(spread); i++ {
		if spread[i].DistanceM < spread[i-1].DistanceM {
			t.Errorf("matches must be ordered by distance: %.1f m before %.1f m",
				spread[i-1].DistanceM, spread[i].DistanceM)
		}
		if spread[i].DistanceM > spread[i-1].DistanceM {
			distinct++
		}
	}
	if distinct == 0 {
		t.Error("every match was the same distance away, so this proved nothing about order")
	}
	for _, m := range spread {
		if m.DistanceM > 250 {
			t.Errorf("a match outside the search radius: %q at %.1f m", m.WorkCode, m.DistanceM)
		}
	}

	// Roughly 400 m away — a different street. It must not match at 25 m,
	// because attributing a defect to the neighbouring street's contract is
	// the failure this system cannot have.
	far, err := db.NearestRoadSegments(ctx, 19.2375, 72.8480, 25, 5)
	if err != nil {
		t.Fatalf("NearestRoadSegments: %v", err)
	}
	for _, m := range far {
		if m.WorkCode == "W-415-test" {
			t.Errorf("a point 400 m away matched at a 25 m buffer (%.1f m)", m.DistanceM)
		}
	}
}
