package store

import "testing"

// A deadline tells a resident what the authority owes them. A process fact
// tells them how the thing in the photograph is lawfully done — what
// permission it needs, what that permission costs, how long the authority has
// to decide. The two must never be confused, so they are selected by separate
// maps and only `tree felling` has the second kind.
func TestProcessFactsAreScopedToWhatHasThem(t *testing.T) {
	if got := processFactsFor("environment", "tree felling"); len(got) != 3 {
		t.Errorf("tree felling has %d process fact(s), want 3", len(got))
	}
	for _, c := range [][2]string{
		{"road_defect", "pothole"},
		{"waste", "dead animal"},
		{"environment", "mangrove destruction"},
		{"environment", ""},
	} {
		if got := processFactsFor(c[0], c[1]); len(got) != 0 {
			t.Errorf("%s/%s invented %d process fact(s)", c[0], c[1], len(got))
		}
	}
}

// Money is stored in paise and shown in rupees, and a figure printed in the
// wrong unit on a public page is the kind of error nobody checks twice.
func TestFormatConstantValue(t *testing.T) {
	cases := []struct {
		key, value, want string
	}{
		{"tree.felling_permission_days", "45", "45 days"},
		{"tree.pruning_decision_days", "7", "7 days"},
		{"tree.felling_processing_fee_paise", "600000", "₹6,000"},
		{"tree.transplant_deposit_paise", "1000000", "₹10,000"},
		{"tree.felling_deposit_paise", "123456", "₹1,234.56"},
		{"something.else", "12", "12"},
		{"tree.felling_permission_days", "not a number", ""},
	}
	for _, c := range cases {
		if got := formatConstantValue(c.key, c.value); got != c.want {
			t.Errorf("formatConstantValue(%q, %q) = %q, want %q",
				c.key, c.value, got, c.want)
		}
	}
}
