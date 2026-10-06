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

// The first two real captures both came back with a subcategory the taxonomy
// has never heard of — "debris / construction and demolition waste dumped on
// roadside", "illegal dumping / uncollected garbage pile". The model was
// describing the photograph rather than choosing from a list, because the
// taxonomy block gave it a comma-joined blob under an instruction about
// categories.
//
// Every such capture becomes needs_confirmation, so a day of walking would
// produce a pile of reports that all need answering by hand and an evaluation
// that measures vocabulary compliance instead of whether the model can see.
func TestTheTaxonomyBlockOffersSubcategoriesAsAClosedList(t *testing.T) {
	block := Prompt{Text: "PROMPT"}.WithTaxonomy()

	// Each subcategory on its own line, so it reads as something to copy
	// rather than prose to paraphrase.
	for _, sub := range []string{"pothole", "open manhole", "illegal dumping"} {
		if !strings.Contains(block, "\n- "+sub+"\n") {
			t.Errorf("%q is not offered as its own item:\n%s", sub, block)
		}
	}

	// And it has to say so in words, about the subcategory specifically.
	lower := strings.ToLower(block)
	if !strings.Contains(lower, "subcategory") {
		t.Error("the block never mentions the subcategory field")
	}
	for _, phrase := range []string{"exactly one", "verbatim"} {
		if !strings.Contains(lower, phrase) {
			t.Errorf("the instruction does not say %q:\n%s", phrase, block)
		}
	}

	// A comma-joined run of subcategories is the shape that caused this.
	if strings.Contains(block, "pothole, crack") {
		t.Error("subcategories are still emitted as a comma-joined blob")
	}
}

// Two subcategories in the taxonomy describe a heap of rubbish, and the
// difference is not visible in the rubbish — it is whether the place is a
// collection point. BMC's own model is that refuse goes to a collection point
// and a compactor empties it on a schedule, so waste there is a failure to
// collect and waste anywhere else is something a person left.
//
// Without the rule stated the model guesses, the labeller guesses differently,
// and the evaluation measures the disagreement rather than the model.
func TestThePromptSeparatesDumpingFromUncollectedWaste(t *testing.T) {
	block := Prompt{Text: "PROMPT"}.WithTaxonomy()
	lower := strings.ToLower(block)

	for _, phrase := range []string{
		"collection point",
		"illegal dumping",
		"uncollected garbage",
	} {
		if !strings.Contains(lower, phrase) {
			t.Errorf("the prompt never mentions %q", phrase)
		}
	}
	// The distinguishing test has to be stated, not implied.
	if !strings.Contains(lower, "bin") {
		t.Error("the prompt does not say what to look for")
	}
}
