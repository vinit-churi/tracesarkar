package jurisdiction

import (
	"encoding/json"
	"fmt"
)

// Boundary is one ward polygon read from BMC's ward layer.
type Boundary struct {
	Ward    string
	GeoJSON string
}

// ParseWardLayer reads BMC's geo/getwardlayer response.
//
// The layer names wards without a separator ("RC"), while the works endpoint
// of the same API uses "R/C". Every code is normalised here so nothing
// downstream has to remember which endpoint it came from.
func ParseWardLayer(body []byte) ([]Boundary, error) {
	var payload struct {
		Features []struct {
			Properties struct {
				WardName string `json:"wardname"`
			} `json:"properties"`
			Geometry json.RawMessage `json:"geometry"`
		} `json:"features"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("ward layer: decode: %w", err)
	}
	if len(payload.Features) == 0 {
		return nil, fmt.Errorf("ward layer: no features")
	}

	out := make([]Boundary, 0, len(payload.Features))
	seen := map[string]bool{}
	for _, f := range payload.Features {
		ward := NormaliseWard(f.Properties.WardName)
		if ward == "" {
			// A polygon we cannot name is a polygon we cannot route with.
			return nil, fmt.Errorf("ward layer: %q is not a Greater Mumbai ward",
				f.Properties.WardName)
		}
		if seen[ward] {
			return nil, fmt.Errorf("ward layer: %s appears twice; the layer no longer has one polygon per ward", ward)
		}
		if len(f.Geometry) == 0 {
			return nil, fmt.Errorf("ward layer: %s has no geometry", ward)
		}
		seen[ward] = true
		out = append(out, Boundary{Ward: ward, GeoJSON: string(f.Geometry)})
	}
	return out, nil
}
