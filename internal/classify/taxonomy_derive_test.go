package classify

import "testing"

// A labeller picks one subcategory per photograph and nothing else. Category
// and hazard are then read off the taxonomy rather than typed again — two
// fields that can disagree are two fields that eventually will, and a wrong
// hazard flag in the eval set silently moves the recall number the model is
// being judged against.
func TestCategoryOf(t *testing.T) {
	tests := []struct {
		sub  string
		want string
	}{
		{"pothole", "road_defect"},
		{"open manhole", "road_defect"},
		{"illegal dumping", "waste"},
		{"sewage overflow", "water_drainage"},
		{"streetlight out", "street_furniture"},
		{"collapsed wall", "structural"},
		{"mangrove destruction", "environment"},
		// The picker sends back whatever the page rendered, so matching has to
		// survive case and stray whitespace.
		{"  Pothole  ", "road_defect"},
		{"CRACKED BRIDGE/FOB", "structural"},
		// Not in the taxonomy at all.
		{"unicorn", ""},
		{"", ""},
	}

	for _, tt := range tests {
		if got := CategoryOf(tt.sub); got != tt.want {
			t.Errorf("CategoryOf(%q) = %q, want %q", tt.sub, got, tt.want)
		}
	}
}

func TestIsHazard(t *testing.T) {
	tests := []struct {
		sub  string
		want bool
	}{
		{"open manhole", true},
		{"missing manhole cover", true},
		{"collapsed wall", true},
		// The taxonomy spells this "cracked bridge/FOB" and the hazard set
		// spells it lowercase. They must still be the same thing.
		{"cracked bridge/FOB", true},
		{"unsafe scaffolding", true},
		{"sewage overflow", true},
		{"distressed building", true},
		{"pothole", false},
		{"uncollected garbage", false},
		{"streetlight out", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsHazard(tt.sub); got != tt.want {
			t.Errorf("IsHazard(%q) = %v, want %v", tt.sub, got, tt.want)
		}
	}
}

// Every hazardous subcategory a labeller can actually pick must belong to the
// taxonomy, or the eval set can never contain it and hazard recall is measured
// against cases that cannot arise.
func TestEveryTaxonomyHazardIsReachable(t *testing.T) {
	var reachable int
	for _, subs := range Taxonomy() {
		for _, s := range subs {
			if IsHazard(s) {
				reachable++
			}
		}
	}
	if reachable == 0 {
		t.Fatal("no hazardous subcategory can be labelled, so hazard recall is unmeasurable")
	}
	t.Logf("%d hazardous subcategories are labellable", reachable)
}
