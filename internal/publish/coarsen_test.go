package publish

import (
	"math"
	"testing"
)

// Borivali, near where the first real captures were taken.
const (
	lat = 19.234093
	lon = 72.844189
)

// Snapping, not jittering. The difference is the whole security property:
// random offsets average away. Publish the same point a hundred times with a
// fresh offset each time and the mean converges on the truth; snap it a
// hundred times and you get the same cell a hundred times.
func TestSnappingIsDeterministic(t *testing.T) {
	first := Coarsen(lat, lon)
	for range 100 {
		again := Coarsen(lat, lon)
		if again != first {
			t.Fatalf("the same point coarsened two ways: %v then %v", first, again)
		}
	}
}

// Two reports from the same block must land on the same cell, or the cell
// itself identifies which of them it was.
func TestNearbyPointsShareACell(t *testing.T) {
	a := Coarsen(lat, lon)
	b := Coarsen(lat+0.00002, lon-0.00003) // a few metres away
	if a != b {
		t.Errorf("points metres apart landed in different cells: %v and %v", a, b)
	}
}

// And points genuinely far apart must not, or the map says nothing.
func TestDistantPointsDoNotShareACell(t *testing.T) {
	a := Coarsen(lat, lon)
	b := Coarsen(lat+0.01, lon) // about 1.1 km north
	if a == b {
		t.Error("points a kilometre apart landed in the same cell")
	}
}

// The published number must not carry more precision than the grid, because a
// trailing digit is the thing that identifies a doorway.
func TestTheOutputCarriesNoMorePrecisionThanTheGrid(t *testing.T) {
	for _, p := range [][2]float64{
		{lat, lon},
		{19.0, 72.0},
		{-19.234093, -72.844189},
		{0, 0},
	} {
		c := Coarsen(p[0], p[1])
		for _, v := range []float64{c.Lat, c.Lon} {
			scaled := v / gridDegrees
			if math.Abs(scaled-math.Round(scaled)) > 1e-9 {
				t.Errorf("%v is not on the %v grid", v, gridDegrees)
			}
		}
	}
}

// The cell must be at least as coarse as it claims. A caller showing "within
// about 110 m" while the cell is 11 m across would be lying in the safe
// direction's opposite.
func TestTheCellIsAtLeastAsCoarseAsItSays(t *testing.T) {
	c := Coarsen(lat, lon)
	if c.CellMetres < 100 {
		t.Errorf("the cell is %v m across, finer than a city block", c.CellMetres)
	}
	// Moving by less than half a cell must not move the answer, which is what
	// makes the cell size an honest statement of how vague it is.
	same := Coarsen(lat+gridDegrees/4, lon)
	if same != c {
		t.Error("a quarter-cell move changed the cell")
	}
}

// Rounding must not be biased towards zero, or coarsened points drift towards
// the equator and the meridian in a way that is itself informative.
func TestRoundingIsSymmetricAboutZero(t *testing.T) {
	north := Coarsen(gridDegrees*0.6, 0)
	south := Coarsen(-gridDegrees*0.6, 0)
	if north.Lat != -south.Lat {
		t.Errorf("rounding is asymmetric: %v and %v", north.Lat, south.Lat)
	}
}
