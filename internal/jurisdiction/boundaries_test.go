package jurisdiction

import (
	"strings"
	"testing"
)

const wardLayerSample = `{"type":"FeatureCollection","features":[
  {"type":"Feature","properties":{"objectid":"1","wardname":"RC"},
   "geometry":{"type":"MultiPolygon","coordinates":[[[[72.84,19.23],[72.85,19.23],[72.85,19.24],[72.84,19.23]]]]}},
  {"type":"Feature","properties":{"objectid":"2","wardname":"RS"},
   "geometry":{"type":"MultiPolygon","coordinates":[[[[72.82,19.20],[72.83,19.20],[72.83,19.21],[72.82,19.20]]]]}}
]}`

func TestWardLayerCodesAreNormalisedOnTheWayIn(t *testing.T) {
	got, err := ParseWardLayer([]byte(wardLayerSample))
	if err != nil {
		t.Fatalf("ParseWardLayer: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("features: got %d, want 2", len(got))
	}
	// "RC" in, "R/C" out — so nothing downstream has to know which BMC
	// endpoint a code came from.
	if got[0].Ward != "R/C" {
		t.Errorf("ward: got %q, want %q", got[0].Ward, "R/C")
	}
	if !strings.Contains(got[0].GeoJSON, "MultiPolygon") {
		t.Errorf("geometry lost: %q", got[0].GeoJSON)
	}
}

func TestAPolygonThatCannotBeNamedIsRefused(t *testing.T) {
	// A polygon we cannot name is one we cannot route with, and a silently
	// skipped ward is a ward whose reports go nowhere.
	bad := strings.Replace(wardLayerSample, `"wardname":"RC"`, `"wardname":"ZZ"`, 1)
	if _, err := ParseWardLayer([]byte(bad)); err == nil {
		t.Fatal("an unrecognised ward code must stop the load")
	}
}

func TestADuplicatedWardIsRefused(t *testing.T) {
	// One polygon per ward. Two means the layer changed shape, and picking
	// either silently would route half the ward's reports to the wrong one.
	dup := strings.Replace(wardLayerSample, `"wardname":"RS"`, `"wardname":"R/C"`, 1)
	if _, err := ParseWardLayer([]byte(dup)); err == nil {
		t.Fatal("a duplicated ward must stop the load")
	}
}

func TestAnEmptyLayerIsRefused(t *testing.T) {
	if _, err := ParseWardLayer([]byte(`{"type":"FeatureCollection","features":[]}`)); err == nil {
		t.Fatal("an empty ward layer must be an error, not zero wards loaded")
	}
}
