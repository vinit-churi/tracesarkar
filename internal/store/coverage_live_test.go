package store_test

import (
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/classify"
)

// Coverage is read from authority_departments at runtime, so a category
// becomes routable the moment a department is recorded for it. That indirection
// is the whole reason adding a category is a migration rather than a code
// change — and it is also why nothing fails loudly if the rows go missing: the
// classifier would simply start answering not_yet_covered for everything, which
// looks like a quiet model problem rather than an empty table.
func TestLiveSeededDepartmentsMakeTheirCategoriesRoutable(t *testing.T) {
	ctx, _, db := liveDB(t)

	got, err := db.CoveredCategories(ctx)
	if err != nil {
		t.Fatalf("CoveredCategories: %v", err)
	}
	have := map[string]bool{}
	for _, c := range got {
		have[c] = true
	}

	// Seeded by migrations 0015 and 0019, each from a first-hand BMC source.
	for _, want := range []string{"road_defect", "waste", "water_drainage"} {
		if !have[want] {
			t.Errorf("%q has no department, so every such capture answers "+
				"not_yet_covered; got %v", want, got)
		}
		// A category that cannot be routed is one the taxonomy should not
		// claim either.
		if !classify.KnownCategory(want) {
			t.Errorf("%q has a department but is not in the taxonomy", want)
		}
	}
}
