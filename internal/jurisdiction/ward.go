// Package jurisdiction resolves a point to the authority that owns it.
//
// Routing a report to the wrong desk is the failure a citizen does not
// forgive, so everything here refuses rather than approximates: an
// unrecognised ward code yields nothing, not a guess.
package jurisdiction

import (
	"sort"
	"strings"
)

// The 24 wards of Greater Mumbai, in the canonical form the glossary fixes
// (docs/00-overview/03-glossary.md). Every dataset is normalised to this.
var wards = map[string]bool{
	"A": true, "B": true, "C": true, "D": true, "E": true,
	"F/N": true, "F/S": true, "G/N": true, "G/S": true,
	"H/E": true, "H/W": true, "K/E": true, "K/W": true,
	"L": true, "M/E": true, "M/W": true, "N": true,
	"P/N": true, "P/S": true, "R/C": true, "R/N": true, "R/S": true,
	"S": true, "T": true,
}

// NormaliseWard turns any of BMC's ward spellings into the canonical one, and
// returns empty for anything that is not a Greater Mumbai ward.
//
// This exists because one BMC API returns both shapes: the works carry "R/C"
// and the boundary layer carries "RC". Joining them without normalising
// matches nothing, silently, for every report — which looks exactly like a
// ward with no problems in it.
func NormaliseWard(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	// Separators seen across BMC's endpoints and spreadsheets.
	s = strings.NewReplacer("-", "", "/", "", " ", "", ".", "").Replace(s)

	switch len(s) {
	case 1:
		if wards[s] {
			return s
		}
	case 2:
		withSlash := s[:1] + "/" + s[1:]
		if wards[withSlash] {
			return withSlash
		}
	}
	return ""
}

// IsKnownWard reports whether a code names a Greater Mumbai ward, in any of
// the spellings NormaliseWard accepts.
func IsKnownWard(raw string) bool { return NormaliseWard(raw) != "" }

// KnownWards lists every ward, canonically, in a stable order.
func KnownWards() []string {
	out := make([]string, 0, len(wards))
	for w := range wards {
		out = append(out, w)
	}
	sort.Strings(out)
	return out
}
