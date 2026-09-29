package roads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type fakeSink struct {
	saved []Segment
	err   error
}

func (f *fakeSink) SaveSegment(_ context.Context, s Segment) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, s)
	return nil
}

func records(t *testing.T, raws ...string) []WorkRecord {
	t.Helper()
	var out []WorkRecord
	for i, raw := range raws {
		var m map[string]any
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatal(err)
		}
		out = append(out, WorkRecord{WorkID: fmt.Sprintf("work-%d", i), Record: m})
	}
	return out
}

const noGeom = `{"workCode":"W-2","wardName":"R/C"}`

func TestLoadStoresEveryWorkThatHasGeometry(t *testing.T) {
	sink := &fakeSink{}
	in := records(t, borivaliWork, borivaliWork)

	result, err := Load(context.Background(), in, sink, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if result.Loaded != 2 {
		t.Errorf("loaded: got %d, want 2", result.Loaded)
	}
	if len(sink.saved) != 2 {
		t.Errorf("sink received %d", len(sink.saved))
	}
}

func TestLoadCountsWorksWithNoGeometryRatherThanFailing(t *testing.T) {
	sink := &fakeSink{}

	// A work with no shape is not an error for the run: BMC simply has not
	// published geometry for it. It has to be counted, though, because a
	// silent skip is how coverage quietly rots.
	result, err := Load(context.Background(), records(t, borivaliWork, noGeom), sink, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if result.Loaded != 1 {
		t.Errorf("loaded: got %d, want 1", result.Loaded)
	}
	if result.NoGeometry != 1 {
		t.Errorf("no-geometry count: got %d, want 1", result.NoGeometry)
	}
}

func TestLoadCanRestrictToOneWard(t *testing.T) {
	sink := &fakeSink{}
	other := `{"workCode":"W-9","wardName":"H/W","geometrytype":{"type":"MultiLineString",
	           "coordinates":[[[72.83,19.05],[72.831,19.051]]]}}`

	result, err := Load(context.Background(), records(t, borivaliWork, other), sink, "R/C")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if result.Loaded != 1 {
		t.Fatalf("loaded: got %d, want 1", result.Loaded)
	}
	if sink.saved[0].Ward != "R/C" {
		t.Errorf("wrong ward loaded: %q", sink.saved[0].Ward)
	}
	if result.OtherWard != 1 {
		t.Errorf("other-ward count: got %d, want 1", result.OtherWard)
	}
}

func TestLoadStopsWhenTheStoreFails(t *testing.T) {
	// A storage failure is not a data-quality skip. Continuing would report a
	// clean run over a half-loaded table.
	sink := &fakeSink{err: errors.New("database is unreachable")}

	if _, err := Load(context.Background(), records(t, borivaliWork), sink, ""); err == nil {
		t.Fatal("a store failure must stop the load")
	}
}

func TestLoadRefusesSuspectCoordinatesLoudly(t *testing.T) {
	swapped := `{"workCode":"W-3","wardName":"R/C","geometrytype":{"type":"MultiLineString",
	             "coordinates":[[[19.234,72.844],[19.233,72.845]]]}}`
	sink := &fakeSink{}

	result, err := Load(context.Background(), records(t, swapped), sink, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Not a quiet no-geometry skip: something is wrong with the source and the
	// count has to say so separately.
	if result.Rejected != 1 {
		t.Errorf("rejected: got %d, want 1", result.Rejected)
	}
	if result.NoGeometry != 0 {
		t.Errorf("a bad coordinate is not a missing one: %+v", result)
	}
}
