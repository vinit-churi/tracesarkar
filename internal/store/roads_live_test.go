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
	near, err := db.NearestRoadSegments(ctx, 19.23391, 72.84440, 25, 5)
	if err != nil {
		t.Fatalf("NearestRoadSegments: %v", err)
	}
	if len(near) == 0 {
		t.Fatal("a point on the road must match it")
	}
	if near[0].WorkCode != "W-415-test" {
		t.Errorf("matched the wrong work: %q", near[0].WorkCode)
	}
	if near[0].DistanceM > 25 {
		t.Errorf("distance should be within the buffer: %v m", near[0].DistanceM)
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
