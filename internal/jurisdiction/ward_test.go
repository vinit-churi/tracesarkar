package jurisdiction

import "testing"

// BMC returns ward codes in two different shapes from two endpoints of the
// same API: the works carry "R/C" and the boundary layer carries "RC". Joining
// them without normalising matches nothing, silently, for every report.
func TestWardCodesFromBothBMCEndpointsNormaliseToTheSameThing(t *testing.T) {
	tests := []struct{ in, want string }{
		// The boundary layer's shape.
		{"RC", "R/C"}, {"RN", "R/N"}, {"RS", "R/S"},
		{"FN", "F/N"}, {"FS", "F/S"}, {"GN", "G/N"}, {"GS", "G/S"},
		{"HE", "H/E"}, {"HW", "H/W"}, {"KE", "K/E"}, {"KW", "K/W"},
		{"ME", "M/E"}, {"MW", "M/W"}, {"PN", "P/N"}, {"PS", "P/S"},
		// Already canonical.
		{"R/C", "R/C"}, {"H/W", "H/W"},
		// Single-letter wards have no sub-division and gain no slash.
		{"A", "A"}, {"B", "B"}, {"L", "L"}, {"N", "N"}, {"S", "S"}, {"T", "T"},
		// Shapes seen in the wild.
		{"r/c", "R/C"}, {" RC ", "R/C"}, {"R-C", "R/C"}, {"R C", "R/C"},
		// Forgiving a doubled separator is safe: stripping separators can
		// never turn one ward into a different one, only into a valid code or
		// into nothing.
		{"R//C", "R/C"},
	}

	for _, tt := range tests {
		if got := NormaliseWard(tt.in); got != tt.want {
			t.Errorf("NormaliseWard(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestEveryWardInGreaterMumbaiIsKnown(t *testing.T) {
	// The glossary's list, which is the canonical one. A ward missing here is
	// a report that cannot be routed.
	canonical := []string{"A", "B", "C", "D", "E", "F/N", "F/S", "G/N", "G/S",
		"H/E", "H/W", "K/E", "K/W", "L", "M/E", "M/W", "N", "P/N", "P/S",
		"R/C", "R/N", "R/S", "S", "T"}

	if len(KnownWards()) != 24 {
		t.Errorf("Greater Mumbai has 24 wards, KnownWards has %d", len(KnownWards()))
	}
	for _, w := range canonical {
		if !IsKnownWard(w) {
			t.Errorf("%q is a Greater Mumbai ward and is not recognised", w)
		}
	}
}

func TestSomethingThatIsNotAWardIsRefused(t *testing.T) {
	// Better an empty answer than a confident wrong one: a report routed to a
	// ward that does not exist goes to nobody.
	for _, in := range []string{"", "ZZ", "R/X", "Borivali", "24", "Q/A", "RCC"} {
		if got := NormaliseWard(in); got != "" {
			t.Errorf("NormaliseWard(%q) = %q, want empty", in, got)
		}
	}
}
