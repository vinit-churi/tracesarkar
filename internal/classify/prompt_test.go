package classify

import (
	"strings"
	"testing"
)

func TestThePromptIsLoadedFromAFileWithItsVersion(t *testing.T) {
	p, err := LoadPrompt("road_defect_v1")
	if err != nil {
		t.Fatalf("LoadPrompt: %v", err)
	}
	if p.Version != "v1" {
		t.Errorf("version: got %q, want %q", p.Version, "v1")
	}
	if strings.HasPrefix(p.Text, "---") {
		t.Error("the front matter should be stripped from the prompt text")
	}
	// Wrapped at 100 columns like every other doc here, so assert on short
	// phrases that cannot straddle a line break, case-insensitively.
	lower := strings.ToLower(p.Text)
	for _, want := range []string{"public infrastructure", "monsoon", "hazard", "rationale"} {
		if !strings.Contains(lower, want) {
			t.Errorf("the prompt does not cover %q", want)
		}
	}
	if len(p.Text) < 1000 {
		t.Errorf("the prompt looks truncated: %d bytes", len(p.Text))
	}
}

func TestTheTaxonomyInThePromptIsTheOneWeValidateAgainst(t *testing.T) {
	// If the prompt and the guardrails drift, the model is told about
	// categories that cannot be routed, and every one of those reports dies
	// at the validation step.
	p, err := LoadPrompt("road_defect_v1")
	if err != nil {
		t.Fatal(err)
	}
	full := p.WithTaxonomy()

	for _, c := range TopCategories() {
		if !strings.Contains(full, c) {
			t.Errorf("the prompt does not mention routable category %q", c)
		}
	}
	if !strings.Contains(full, "pothole") {
		t.Error("subcategories should reach the prompt too")
	}
}

func TestAMissingPromptIsAnErrorNotAnEmptyString(t *testing.T) {
	if _, err := LoadPrompt("does_not_exist"); err == nil {
		t.Fatal("a missing prompt must fail loudly; an empty system prompt would silently degrade every classification")
	}
}
