// Package publish turns what the platform knows into what it may show.
//
// Hard rule 4: precise reporter location never reaches a public surface. The
// exact position stays internal, where it is needed to match a contract and to
// address a complaint; what goes out is a cell.
package publish

import "math"

// gridDegrees is the cell size, in degrees.
//
// A thousandth of a degree is about 111 m north–south, and about 105 m
// east–west at Mumbai's latitude — roughly a city block. Fine enough to say
// which stretch of road a problem is on, coarse enough that it does not say
// which doorway it is outside.
const gridDegrees = 0.001

// metresPerDegree is the north–south figure, used to describe the cell. It is
// the larger of the two dimensions at these latitudes, so describing the cell
// with it overstates rather than understates how vague the point is.
const metresPerDegree = 111320

// Point is a position safe to publish.
type Point struct {
	Lat, Lon float64
	// CellMetres is how wide the cell is, so a surface can say how vague the
	// point is instead of drawing a dot that implies precision.
	CellMetres float64
}

// Coarsen snaps a position to the grid.
//
// Snapped, never jittered. A random offset looks private and is not: publish
// the same point enough times with a fresh offset each time and the mean
// converges on the truth. Snapping gives the same cell every time, so
// repetition reveals nothing that one observation did not.
func Coarsen(lat, lon float64) Point {
	return Point{
		Lat:        snap(lat),
		Lon:        snap(lon),
		CellMetres: gridDegrees * metresPerDegree,
	}
}

// snap rounds to the nearest cell, away from zero on a tie so that northern
// and southern points are treated alike — rounding towards zero would pull
// coarsened positions towards the equator, which is itself a signal.
func snap(v float64) float64 {
	return math.Round(v/gridDegrees) * gridDegrees
}
