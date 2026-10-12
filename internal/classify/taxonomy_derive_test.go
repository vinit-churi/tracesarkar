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

// A manhole cover that is cracked through is neither missing nor open, and
// until it was listed a surveyor standing over one had to choose a label that
// was wrong or skip the photograph. Both poison the evaluation set — the first
// with a bad answer, the second with a hole in it.
func TestADamagedManholeCoverIsLabellableAndHazardous(t *testing.T) {
	if got := CategoryOf("damaged manhole cover"); got != "road_defect" {
		t.Errorf("CategoryOf = %q, want road_defect", got)
	}
	// The hazard is the same whichever way the cover failed: a foot or a
	// wheel goes in.
	if !IsHazard("damaged manhole cover") {
		t.Error("a cover with a hole in it is not treated as a hazard")
	}
}

// Every hazardous subcategory must be one a person can actually choose. A
// hazard the taxonomy knows and the picker does not is a hazard that never
// reaches the evaluation set, and hazard recall is the figure held to 99%.
func TestEveryHazardIsInTheTaxonomy(t *testing.T) {
	for _, h := range Hazards() {
		if CategoryOf(h) == "" {
			t.Errorf("%q is treated as a hazard but belongs to no category", h)
		}
	}
	if len(Hazards()) < 8 {
		t.Errorf("only %d labellable hazards", len(Hazards()))
	}
}
