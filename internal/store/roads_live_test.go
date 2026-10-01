package store_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
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

// Two roads under different contracts can occupy the same place: 269 pairs of
// segments from different contract packages lie within a metre of each other in
// the collected data. The attribution gate handles that honestly — it spots a
// rival giving a different answer and asks rather than guessing — but it reads
// candidates[0] as the nearest, and with no tie-break the database is free to
// pick either.
//
// That would make the contractor named against a photograph depend on the query
// plan rather than on the data, and re-running enrichment could name a
// different company for the same capture, with nothing having changed. A fact
// published about a named party has to be reproducible.
func TestLiveTiedRoadSegmentsComeBackInADefinedOrder(t *testing.T) {
	ctx, pool, db := liveDB(t)

	var workIDs []string
	rows, err := pool.Query(ctx, `SELECT id FROM works LIMIT 2`)
	if err != nil {
		t.Skipf("no collected works to project from: %v", err)
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		workIDs = append(workIDs, id)
	}
	rows.Close()
	if len(workIDs) < 2 {
		t.Skip("need two collected works to build a tie")
	}

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM road_segments WHERE work_code LIKE 'TIE-%'`)
	})

	// Inserted in descending id order, deliberately. Without a tie-break the
	// rows come back in whatever order the scan produces, which is insertion
	// order — so inserting in the order the test requires would let it pass by
	// luck, which is exactly what it did before this line existed.
	sort.Sort(sort.Reverse(sort.StringSlice(workIDs)))

	// Identical geometry, different contracts — the shape of the 269 real pairs.
	// Somewhere in the Arabian Sea off Borivali, so no collected segment can
	// join the tie and change what is being measured.
	geo := `{"type":"MultiLineString","coordinates":[[` +
		`[72.70000,19.23000],[72.70100,19.23010]]]}`
	for i, id := range workIDs {
		if err := db.SaveRoadSegment(ctx, store.NewRoadSegment{
			WorkID:       id,
			WorkCode:     fmt.Sprintf("TIE-%d", i),
			Ward:         "R/C",
			LocationName: fmt.Sprintf("Tied stretch %d", i),
			GeoJSON:      geo,
		}); err != nil {
			t.Fatalf("SaveRoadSegment %d: %v", i, err)
		}
	}

	var first []string
	for call := range 3 {
		got, err := db.NearestRoadSegments(ctx, 19.23005, 72.70050, 50, 5)
		if err != nil {
			t.Fatalf("NearestRoadSegments: %v", err)
		}
		var order []string
		for _, m := range got {
			if strings.HasPrefix(m.WorkCode, "TIE-") {
				order = append(order, m.WorkID)
			}
		}
		if len(order) != 2 {
			t.Fatalf("both tied segments must be returned; got %d", len(order))
		}
		if call == 0 {
			first = order
			// The rule itself, not merely that it is stable: an order that is
			// consistent only because the plan has not changed is not a
			// guarantee of anything.
			if order[0] >= order[1] {
				t.Errorf("ties must break on a defined key; got %v", order)
			}
			continue
		}
		if order[0] != first[0] || order[1] != first[1] {
			t.Errorf("call %d returned a different order: %v then %v", call, first, order)
		}
	}
}
