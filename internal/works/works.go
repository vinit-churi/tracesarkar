// Package works parses the works datasets that authorities publish, and
// compares successive snapshots.
//
// Two rules hold for every parser here:
//   - fields on the source's register blocklist are dropped before a record
//     exists, so personal data never reaches the database;
//   - each endpoint has one fixed natural key, and a key collision is an error
//     rather than a silent merge.
package works

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Record is one row of a works dataset after the blocklist strip.
type Record struct {
	NaturalKey string
	Fields     map[string]any
	// Vanished marks a record that was absent from the previous snapshot; it is
	// set when loading prior state, not by the parsers.
	Vanished bool
}

// Parser turns a response body into records.
type Parser func(body []byte, blocklist []string) ([]Record, error)

// ParseRoadsDashboard reads BMC's `publicdashboard/` response. The list carries
// no identifier, so the key is road name, ward and contractor together.
func ParseRoadsDashboard(body []byte, blocklist []string) ([]Record, error) {
	var payload struct {
		Count int              `json:"count"`
		Data  []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("roads dashboard: decode: %w", err)
	}
	if payload.Data == nil {
		return nil, fmt.Errorf("roads dashboard: response has no data array")
	}
	return buildRecords(payload.Data, blocklist, func(row map[string]any) string {
		return joinKey(str(row["roadName"]), str(row["ward"]), str(row["contractorName"]))
	})
}

// ParseRoadsGeometry reads the GeoJSON road layer, whose nested location object
// carries the work code.
func ParseRoadsGeometry(body []byte, blocklist []string) ([]Record, error) {
	var payload struct {
		Features []struct {
			Properties map[string]any  `json:"properties"`
			Geometry   json.RawMessage `json:"geometry"`
		} `json:"features"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("roads geometry: decode: %w", err)
	}
	if payload.Features == nil {
		return nil, fmt.Errorf("roads geometry: response has no features array")
	}
	rows := make([]map[string]any, 0, len(payload.Features))
	for _, f := range payload.Features {
		row := map[string]any{}
		for k, v := range f.Properties {
			row[k] = v
		}
		// The location object holds the work code and dates.
		if loc, ok := f.Properties["location"].(map[string]any); ok {
			for k, v := range loc {
				row[k] = v
			}
			delete(row, "location")
		}
		rows = append(rows, row)
	}
	return buildRecords(rows, blocklist, func(row map[string]any) string {
		return joinKey(str(row["workCode"]), str(row["locationName"]))
	})
}

// ParseWardMaster reads the ward master list.
func ParseWardMaster(body []byte, blocklist []string) ([]Record, error) {
	var payload struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("ward master: decode: %w", err)
	}
	if payload.Data == nil {
		return nil, fmt.Errorf("ward master: response has no data array")
	}
	return buildRecords(payload.Data, blocklist, func(row map[string]any) string {
		return joinKey(str(row["wardID"]), str(row["wardName"]))
	})
}

// ParseSWDProgressCard reads the storm-water drain progress card, which is a
// single object describing the season so far.
func ParseSWDProgressCard(body []byte, blocklist []string) ([]Record, error) {
	var payload struct {
		Data map[string]any `json:"Data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("swd progress card: decode: %w", err)
	}
	if payload.Data == nil {
		return nil, fmt.Errorf("swd progress card: response has no Data object")
	}
	return buildRecords([]map[string]any{payload.Data}, blocklist, func(map[string]any) string {
		return "progresscard"
	})
}

// ParseSWDRows reads the storm-water endpoints that return an array.
func ParseSWDRows(keyFields ...string) Parser {
	return func(body []byte, blocklist []string) ([]Record, error) {
		var payload struct {
			Data []map[string]any `json:"Data"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("swd rows: decode: %w", err)
		}
		if payload.Data == nil {
			return nil, fmt.Errorf("swd rows: response has no Data array")
		}
		return buildRecords(payload.Data, blocklist, func(row map[string]any) string {
			parts := make([]string, 0, len(keyFields))
			for _, f := range keyFields {
				parts = append(parts, str(row[f]))
			}
			return joinKey(parts...)
		})
	}
}

func buildRecords(rows []map[string]any, blocklist []string, key func(map[string]any) string) ([]Record, error) {
	banned := make(map[string]bool, len(blocklist))
	for _, f := range blocklist {
		banned[strings.ToLower(strings.TrimSpace(f))] = true
	}

	out := make([]Record, 0, len(rows))
	seen := make(map[string]bool, len(rows))
	for i, row := range rows {
		fields := make(map[string]any, len(row))
		for k, v := range row {
			if banned[strings.ToLower(k)] {
				continue
			}
			fields[k] = v
		}
		k := key(fields)
		if strings.TrimSpace(strings.ReplaceAll(k, "|", "")) == "" {
			return nil, fmt.Errorf("record %d has an empty natural key", i)
		}
		if seen[k] {
			return nil, fmt.Errorf("natural key %q appears twice; the key for this endpoint no longer identifies a record", k)
		}
		seen[k] = true
		out = append(out, Record{NaturalKey: k, Fields: fields})
	}
	return out, nil
}

func joinKey(parts ...string) string {
	trimmed := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed = append(trimmed, strings.TrimSpace(p))
	}
	return strings.Join(trimmed, "|")
}

func str(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%f", t), "0"), ".")
	default:
		return fmt.Sprintf("%v", t)
	}
}
