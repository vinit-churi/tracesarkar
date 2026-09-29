package roads

import (
	"encoding/json"
	"math"
	"testing"
)

// A real record, as BMC's roads API returns it: the geometry lives under
// "geometrytype", and "coordinates" is an empty string array that carries
// nothing. Reading the wrong one silently yields no geometry at all.
const borivaliWork = `{
  "workCode": "W-415",
  "locationName": "Derasar to Dead",
  "coordinates": [""],
  "geometrytype": {
    "type": "MultiLineString",
    "coordinates": [[
      [72.8441898127395, 19.234093217066558],
      [72.84440287106771, 19.233907857530042],
      [72.84478303413812, 19.233418822942667]
    ]]
  },
  "wardName": "R/C",
  "startDate": "2024-01-09T00:00:00.000Z",
  "endDate": "2024-05-31T00:00:00.000Z"
}`

func recordFrom(t *testing.T, raw string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestSegmentReadsGeometryFromTheFieldThatHasIt(t *testing.T) {
	seg, err := SegmentFrom(recordFrom(t, borivaliWork))
	if err != nil {
		t.Fatalf("SegmentFrom: %v", err)
	}

	if seg.WorkCode != "W-415" {
		t.Errorf("work code: %q", seg.WorkCode)
	}
	if seg.Ward != "R/C" {
		t.Errorf("ward: %q", seg.Ward)
	}
	if seg.LocationName != "Derasar to Dead" {
		t.Errorf("location: %q", seg.LocationName)
	}
	if len(seg.Lines) != 1 || len(seg.Lines[0]) != 3 {
		t.Fatalf("expected one line of three points, got %d lines", len(seg.Lines))
	}

	// Longitude first, as GeoJSON specifies. Swapping them puts Mumbai in the
	// Indian Ocean and every join silently returns nothing.
	first := seg.Lines[0][0]
	if math.Abs(first.Lon-72.8441898127395) > 1e-9 {
		t.Errorf("lon: got %v", first.Lon)
	}
	if math.Abs(first.Lat-19.234093217066558) > 1e-9 {
		t.Errorf("lat: got %v", first.Lat)
	}
}

func TestSegmentRejectsAWorkWithNoGeometry(t *testing.T) {
	for name, raw := range map[string]string{
		"no geometrytype":     `{"workCode":"W-1","wardName":"R/C"}`,
		"empty coordinates":   `{"workCode":"W-1","wardName":"R/C","geometrytype":{"type":"MultiLineString","coordinates":[]}}`,
		"only the empty list": `{"workCode":"W-1","wardName":"R/C","coordinates":[""]}`,
		"line with one point": `{"workCode":"W-1","wardName":"R/C","geometrytype":{"type":"MultiLineString","coordinates":[[[72.8,19.2]]]}}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := SegmentFrom(recordFrom(t, raw)); err == nil {
				t.Error("a work with no usable geometry must be refused, not stored as empty")
			}
		})
	}
}

func TestSegmentRefusesCoordinatesOutsideTheRegion(t *testing.T) {
	// A swapped lat/lon pair is the classic silent failure: it parses, it
	// stores, and every spatial join afterwards returns nothing.
	swapped := `{"workCode":"W-2","wardName":"R/C","geometrytype":{"type":"MultiLineString",
	             "coordinates":[[[19.234,72.844],[19.233,72.845]]]}}`
	if _, err := SegmentFrom(recordFrom(t, swapped)); err == nil {
		t.Error("coordinates outside the Mumbai region must be refused")
	}
}

func TestSegmentAsGeoJSONRoundTrips(t *testing.T) {
	seg, err := SegmentFrom(recordFrom(t, borivaliWork))
	if err != nil {
		t.Fatal(err)
	}

	// PostGIS ingests this via ST_GeomFromGeoJSON, so it has to be valid
	// GeoJSON with the coordinates in the order GeoJSON expects.
	var back struct {
		Type        string        `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"`
	}
	if err := json.Unmarshal([]byte(seg.GeoJSON()), &back); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if back.Type != "MultiLineString" {
		t.Errorf("type: %q", back.Type)
	}
	if got := back.Coordinates[0][0]; got[0] != 72.8441898127395 || got[1] != 19.234093217066558 {
		t.Errorf("first point: %v", got)
	}
}

func TestSegmentKeepsTheDatesThatBoundAWarranty(t *testing.T) {
	seg, err := SegmentFrom(recordFrom(t, borivaliWork))
	if err != nil {
		t.Fatal(err)
	}
	if seg.EndDate == nil {
		t.Fatal("the completion date bounds the defect liability period; it must survive")
	}
	if seg.EndDate.Format("2006-01-02") != "2024-05-31" {
		t.Errorf("end date: %s", seg.EndDate)
	}
}

func TestSegmentReadsTheContractorAndTrimsTheCodeBMCAppends(t *testing.T) {
	raw := `{"workCode":"W-415","wardName":"R/C",
	         "contractorName":"M/s Dineshchandra Ramchandra Agrawal Infracon Pvt. Ltd. W-415.",
	         "geometrytype":{"type":"MultiLineString","coordinates":[[[72.844,19.234],[72.845,19.233]]]}}`
	seg, err := SegmentFrom(recordFrom(t, raw))
	if err != nil {
		t.Fatal(err)
	}
	// The package code is not part of the company's name, and it will be shown
	// to people as the party responsible.
	want := "M/s Dineshchandra Ramchandra Agrawal Infracon Pvt. Ltd"
	if seg.ContractorName != want {
		t.Errorf("contractor: got %q, want %q", seg.ContractorName, want)
	}
}

func TestSegmentWithNoContractorLeavesItEmptyRatherThanGuessing(t *testing.T) {
	seg, err := SegmentFrom(recordFrom(t, borivaliWork))
	if err != nil {
		t.Fatal(err)
	}
	if seg.ContractorName != "" {
		t.Errorf("an absent contractor must stay absent, got %q", seg.ContractorName)
	}
}
