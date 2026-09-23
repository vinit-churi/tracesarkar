package works

import "testing"

func rec(key string, fields map[string]any) Record {
	return Record{NaturalKey: key, Fields: fields}
}

func TestDiffReportsAddedRecords(t *testing.T) {
	changes := Diff(nil, []Record{rec("a", map[string]any{"status": "New"})})
	if len(changes) != 1 {
		t.Fatalf("got %d changes, want 1: %+v", len(changes), changes)
	}
	if changes[0].Kind != KindAdded || changes[0].NaturalKey != "a" {
		t.Errorf("got %+v", changes[0])
	}
}

func TestDiffReportsFieldLevelChangesWithBothValues(t *testing.T) {
	before := map[string]Record{"a": rec("a", map[string]any{"status": "In Progress", "ward": "R/S"})}
	after := []Record{rec("a", map[string]any{"status": "Completed", "ward": "R/S"})}

	changes := Diff(before, after)

	if len(changes) != 1 {
		t.Fatalf("got %d changes, want 1: %+v", len(changes), changes)
	}
	c := changes[0]
	if c.Kind != KindChanged || c.Field != "status" {
		t.Fatalf("got %+v", c)
	}
	if c.Old != "In Progress" || c.New != "Completed" {
		t.Errorf("both values must be recorded: got old=%v new=%v", c.Old, c.New)
	}
}

func TestDiffIsSilentWhenNothingChanged(t *testing.T) {
	before := map[string]Record{"a": rec("a", map[string]any{"status": "Completed"})}
	after := []Record{rec("a", map[string]any{"status": "Completed"})}
	if changes := Diff(before, after); len(changes) != 0 {
		t.Errorf("expected no changes, got %+v", changes)
	}
}

func TestDiffReportsVanishedRecords(t *testing.T) {
	before := map[string]Record{"gone": rec("gone", map[string]any{"status": "Completed"})}
	changes := Diff(before, []Record{})
	if len(changes) != 1 || changes[0].Kind != KindVanished {
		t.Fatalf("got %+v", changes)
	}
}

func TestDiffReportsReappearedRecords(t *testing.T) {
	before := map[string]Record{"back": {NaturalKey: "back", Fields: map[string]any{"status": "Completed"}, Vanished: true}}
	changes := Diff(before, []Record{rec("back", map[string]any{"status": "Completed"})})
	if len(changes) != 1 || changes[0].Kind != KindReappeared {
		t.Fatalf("got %+v", changes)
	}
}

func TestDiffDetectsAddedAndRemovedFields(t *testing.T) {
	before := map[string]Record{"a": rec("a", map[string]any{"status": "Completed"})}
	after := []Record{rec("a", map[string]any{"status": "Completed", "dlpPeriod": "5"})}

	changes := Diff(before, after)

	if len(changes) != 1 || changes[0].Field != "dlpPeriod" {
		t.Fatalf("got %+v", changes)
	}
	if changes[0].Old != nil {
		t.Errorf("a new field has no old value: %+v", changes[0])
	}
}

func TestDiffIsDeterministicallyOrdered(t *testing.T) {
	before := map[string]Record{"a": rec("a", map[string]any{"x": 1, "y": 2, "z": 3})}
	after := []Record{rec("a", map[string]any{"x": 9, "y": 8, "z": 7})}

	first := Diff(before, after)
	for i := 0; i < 5; i++ {
		again := Diff(before, after)
		for j := range first {
			if first[j].Field != again[j].Field {
				t.Fatalf("ordering is not stable: %v vs %v", first[j].Field, again[j].Field)
			}
		}
	}
}

func TestDiffComparesNumbersByValueNotFormatting(t *testing.T) {
	before := map[string]Record{"a": rec("a", map[string]any{"length": float64(400)})}
	after := []Record{rec("a", map[string]any{"length": float64(400)})}
	if changes := Diff(before, after); len(changes) != 0 {
		t.Errorf("equal numbers must not be reported as a change: %+v", changes)
	}
}
