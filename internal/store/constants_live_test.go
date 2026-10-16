package store_test

import (
	"strings"
	"testing"
)

// The Tree Authority's clocks and fees come from a notification under the
// Maharashtra Right to Public Services Act 2015 — the first statutory service
// deadlines this platform holds, as opposed to a court order, a tender clause
// or a departmental charter. Nothing displays them yet, so the thing that
// stops them rotting is that code can read them and a test asserts what they
// say.
func TestLiveTheTreeAuthorityConstantsAreReadable(t *testing.T) {
	ctx, _, db := liveDB(t)

	cases := []struct {
		key  string
		want string
	}{
		{"tree.pruning_decision_days", "7"},
		{"tree.felling_permission_days", "45"},
		{"tree.felling_processing_fee_paise", "600000"},
		{"tree.felling_deposit_paise", "600000"},
		{"tree.transplant_deposit_paise", "1000000"},
	}
	for _, c := range cases {
		got, found, err := db.LegalConstant(ctx, c.key)
		if err != nil {
			t.Fatalf("LegalConstant(%s): %v", c.key, err)
		}
		if !found {
			t.Errorf("%s is not recorded", c.key)
			continue
		}
		if got.Value != c.want {
			t.Errorf("%s = %q, want %q", c.key, got.Value, c.want)
		}
		// Hard rule 6's other half: a constant is only as good as its citation.
		if !strings.Contains(got.Citation, "Right to Public Services") {
			t.Errorf("%s does not cite the statute: %q", c.key, got.Citation)
		}
		if got.ShortCitation == "" {
			t.Errorf("%s has no short citation, so it can never reach a post", c.key)
		}
	}
}

// A clock for granting permission to fell a tree is not a clock for answering
// a report that one was felled illegally. The platform holds both facts and
// must not confuse them: no deadline is attached to the `tree felling`
// subcategory, and the routing lookup must keep saying so.
func TestLiveReportingAFelledTreeCarriesNoDeadline(t *testing.T) {
	ctx, _, db := liveDB(t)

	r, found, err := db.RoutingFor(ctx, "BMC", "R/C", "environment", "tree felling")
	if err != nil || !found {
		t.Fatalf("RoutingFor: %v (found=%v)", err, found)
	}
	if r.DeadlineHours != 0 {
		t.Errorf("a report of illegal felling was given %d hours; the 45 days are "+
			"for granting permission, not for answering a complaint", r.DeadlineHours)
	}
}
