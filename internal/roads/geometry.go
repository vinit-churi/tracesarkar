// Package roads turns BMC's road-works records into road geometry the
// attribution join can use.
//
// The join is the platform's central claim — this stretch is covered by this
// contract — so everything here refuses rather than guesses. A work with no
// usable geometry is an error, not a row with an empty shape, because a silent
// empty shape matches nothing and looks exactly like a road that simply has no
// defects reported on it.
package roads

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Point is a WGS84 coordinate, in GeoJSON order.
type Point struct {
	Lon float64
	Lat float64
}

// Segment is one road work with the shape of the road it covers.
type Segment struct {
	// WorkID is the archive row this was projected from, so a segment can
	// always be traced back to the document it came from.
	WorkID       string
	WorkCode     string
	Ward         string
	LocationName string
	Lines        [][]Point
	StartDate    *time.Time
	EndDate      *time.Time
}

// The Mumbai Metropolitan Region, generously bounded. This exists to catch the
// classic failure: latitude and longitude swapped. That parses, stores, and
// then silently matches nothing forever.
const (
	minLon, maxLon = 72.0, 73.7
	minLat, maxLat = 18.5, 20.0
)

var errNoGeometry = errors.New("work has no usable geometry")

// SegmentFrom reads one work record.
//
// The geometry is under "geometrytype", not "coordinates" — BMC's API carries
// an empty string array in the field whose name suggests it holds the shape.
func SegmentFrom(record map[string]any) (Segment, error) {
	seg := Segment{
		WorkCode:     text(record["workCode"]),
		Ward:         ward(record),
		LocationName: text(record["locationName"]),
		StartDate:    date(record["startDate"]),
		EndDate:      date(record["endDate"]),
	}

	geometry, ok := record["geometrytype"].(map[string]any)
	if !ok {
		return Segment{}, fmt.Errorf("%s: %w", seg.WorkCode, errNoGeometry)
	}
	rawLines, ok := geometry["coordinates"].([]any)
	if !ok || len(rawLines) == 0 {
		return Segment{}, fmt.Errorf("%s: %w", seg.WorkCode, errNoGeometry)
	}

	for _, rawLine := range rawLines {
		points, ok := rawLine.([]any)
		if !ok {
			continue
		}
		// A line needs two points to have direction or length; one point is a
		// dot, and a buffer around it would attribute a whole junction.
		if len(points) < 2 {
			continue
		}

		line := make([]Point, 0, len(points))
		for _, rawPoint := range points {
			pair, ok := rawPoint.([]any)
			if !ok || len(pair) < 2 {
				continue
			}
			lon, lonOK := pair[0].(float64)
			lat, latOK := pair[1].(float64)
			if !lonOK || !latOK {
				continue
			}
			if lon < minLon || lon > maxLon || lat < minLat || lat > maxLat {
				return Segment{}, fmt.Errorf(
					"%s: point (%.6f, %.6f) is outside the Mumbai region; "+
						"latitude and longitude may be swapped", seg.WorkCode, lon, lat)
			}
			line = append(line, Point{Lon: lon, Lat: lat})
		}
		if len(line) >= 2 {
			seg.Lines = append(seg.Lines, line)
		}
	}

	if len(seg.Lines) == 0 {
		return Segment{}, fmt.Errorf("%s: %w", seg.WorkCode, errNoGeometry)
	}
	return seg, nil
}

// GeoJSON renders the segment for ST_GeomFromGeoJSON.
func (s Segment) GeoJSON() string {
	var b strings.Builder
	b.WriteString(`{"type":"MultiLineString","coordinates":[`)
	for i, line := range s.Lines {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('[')
		for j, p := range line {
			if j > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, "[%v,%v]", p.Lon, p.Lat)
		}
		b.WriteByte(']')
	}
	b.WriteString("]}")
	return b.String()
}

// ward reads the ward name, which BMC returns either as a plain string or as
// the whole ward record nested under the same key.
func ward(record map[string]any) string {
	switch v := record["wardName"].(type) {
	case string:
		return v
	case map[string]any:
		return text(v["wardName"])
	}
	return ""
}

func text(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func date(v any) *time.Time {
	s := text(v)
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

// ErrNoGeometry reports whether a work was skipped for want of a shape.
func ErrNoGeometry(err error) bool { return errors.Is(err, errNoGeometry) }

