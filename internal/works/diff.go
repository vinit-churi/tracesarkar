package works

import (
	"encoding/json"
	"sort"
)

// Change kinds.
const (
	KindAdded      = "added"
	KindChanged    = "changed"
	KindVanished   = "vanished"
	KindReappeared = "reappeared"
)

// Change is one observed difference between two snapshots. A changed field
// carries both values, because the point of the archive is that the previous
// value is recoverable.
type Change struct {
	Kind       string
	NaturalKey string
	Field      string
	Old        any
	New        any
}

// Diff compares the previous state of a dataset with a fresh snapshot.
// Ordering is deterministic: by natural key, then by field name.
func Diff(previous map[string]Record, current []Record) []Change {
	var changes []Change

	seen := make(map[string]bool, len(current))
	sorted := append([]Record(nil), current...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].NaturalKey < sorted[j].NaturalKey })

	for _, rec := range sorted {
		seen[rec.NaturalKey] = true
		before, existed := previous[rec.NaturalKey]
		if !existed {
			changes = append(changes, Change{Kind: KindAdded, NaturalKey: rec.NaturalKey})
			continue
		}
		if before.Vanished {
			changes = append(changes, Change{Kind: KindReappeared, NaturalKey: rec.NaturalKey})
		}
		for _, field := range unionOfKeys(before.Fields, rec.Fields) {
			oldValue, newValue := before.Fields[field], rec.Fields[field]
			if equalValues(oldValue, newValue) {
				continue
			}
			changes = append(changes, Change{
				Kind:       KindChanged,
				NaturalKey: rec.NaturalKey,
				Field:      field,
				Old:        oldValue,
				New:        newValue,
			})
		}
	}

	var gone []string
	for key, rec := range previous {
		if !seen[key] && !rec.Vanished {
			gone = append(gone, key)
		}
	}
	sort.Strings(gone)
	for _, key := range gone {
		changes = append(changes, Change{Kind: KindVanished, NaturalKey: key})
	}

	return changes
}

func unionOfKeys(a, b map[string]any) []string {
	set := make(map[string]bool, len(a)+len(b))
	for k := range a {
		set[k] = true
	}
	for k := range b {
		set[k] = true
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// equalValues compares by JSON representation, which treats numbers by value
// and handles nested objects without reflection surprises.
func equalValues(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	ja, errA := json.Marshal(a)
	jb, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	return string(ja) == string(jb)
}
