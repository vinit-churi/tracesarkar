package classify

import (
	"strings"
	"testing"
)

// allCovered is the coverage a fully built-out platform would have. Tests
// about guardrails should not also be testing coverage.
func allCovered() Coverage { return CoverageFor(TopCategories()...) }

func pothole() Result {
	return Result{
		Category: "road_defect", Subcategory: "pothole",
		Severity: "high", IsCivicIssue: true, ImageQuality: "good",
		Confidence: 0.91,
	}
}

func TestAKnownHazardIsNeverLeftToTheModel(t *testing.T) {
	// An open manhole is lethal. If the model says it is not a hazard, the
	// model is overruled — a false negative here is a person in a hole.
	for _, sub := range []string{"open manhole", "missing manhole cover", "live wire", "collapsed wall"} {
		r := pothole()
		r.Subcategory = sub
		r.HazardToLife = false

		got := Apply(r, allCovered())

		if !got.HazardToLife {
			t.Errorf("%q must be a hazard regardless of what the model said", sub)
		}
		if !got.Overridden {
			t.Errorf("%q: the override must be recorded, not silent", sub)
		}
	}
}

func TestTheModelMaySaySomethingIsAHazardEvenWhenTheListDoesNot(t *testing.T) {
	// The list is a floor, not a ceiling.
	r := pothole()
	r.HazardToLife = true

	got := Apply(r, allCovered())

	if !got.HazardToLife {
		t.Error("a model-reported hazard must survive")
	}
	if got.Overridden {
		t.Error("nothing was overridden; the flag should be false")
	}
}

func TestAnUnusableImageAsksForARetakeRatherThanGuessing(t *testing.T) {
	r := pothole()
	r.ImageQuality = "unusable"

	got := Apply(r, allCovered())

	if got.Outcome != NeedsRetake {
		t.Errorf("outcome: got %q, want %q", got.Outcome, NeedsRetake)
	}
}

func TestSomethingOutOfScopeIsRejectedSpecifically(t *testing.T) {
	r := pothole()
	r.IsCivicIssue = false
	r.Category = "private_property"

	got := Apply(r, allCovered())

	if got.Outcome != OutOfScope {
		t.Errorf("outcome: got %q, want %q", got.Outcome, OutOfScope)
	}
	// Hard rule: the rejection names what it thinks this is. A generic error
	// is what every predecessor platform does and it is why people stop.
	if got.RejectionReason == "" {
		t.Error("an out-of-scope result must carry a specific reason")
	}
}

func TestLowConfidenceAsksTheCitizenRatherThanCommitting(t *testing.T) {
	r := pothole()
	r.Confidence = 0.42
	r.Alternatives = []string{"pothole", "utility-dig damage", "subsidence"}

	got := Apply(r, allCovered())

	if got.Outcome != NeedsConfirmation {
		t.Errorf("outcome: got %q, want %q", got.Outcome, NeedsConfirmation)
	}
	if len(got.Alternatives) == 0 {
		t.Error("the citizen needs options to choose between")
	}
}

func TestHighConfidenceOnAKnownCategoryIsAccepted(t *testing.T) {
	got := Apply(pothole(), allCovered())

	if got.Outcome != Accepted {
		t.Errorf("outcome: got %q, want %q", got.Outcome, Accepted)
	}
}

func TestACategoryOutsideTheTaxonomyIsNotAccepted(t *testing.T) {
	// Structured output guarantees the shape, not the vocabulary. A category
	// we do not have a department for cannot be routed.
	r := pothole()
	r.Category = "alien_invasion"

	got := Apply(r, allCovered())

	if got.Outcome == Accepted {
		t.Errorf("an unknown category must not be accepted: %+v", got)
	}
}

func TestPeoplePresentForcesRedactionBeforeAnythingPublic(t *testing.T) {
	r := pothole()
	r.PeoplePresent = true

	got := Apply(r, allCovered())

	if !got.RedactionRequired {
		t.Error("a photograph with people in it cannot produce a public derivative unredacted")
	}
}

func TestTheRejectionNeverBlamesTheCitizen(t *testing.T) {
	r := pothole()
	r.IsCivicIssue = false
	r.Category = "private_property"

	got := Apply(r, allCovered())

	lower := strings.ToLower(got.RejectionReason)
	for _, word := range []string{"invalid", "wrong", "failed", "error", "you should"} {
		if strings.Contains(lower, word) {
			t.Errorf("the rejection reads as a rebuke (%q): %q", word, got.RejectionReason)
		}
	}
}

func TestSomethingWeRecogniseButCannotRouteIsSaidPlainly(t *testing.T) {
	// v1 routes road defects in one ward. A photograph of uncollected
	// garbage is a real civic problem and we know what it is — but there is
	// no department row behind `waste`, so calling it "accepted" would mean
	// classifying a report and then dropping it. The citizen would see it
	// recognised and never hear anything again.
	r := pothole()
	r.Category = "waste"
	r.Subcategory = "uncollected garbage"

	got := Apply(r, CoverageFor("road_defect"))

	if got.Outcome != NotYetCovered {
		t.Errorf("outcome: got %q, want %q", got.Outcome, NotYetCovered)
	}
	if got.RejectionReason == "" {
		t.Error("the citizen must be told what this is and that it is not covered yet")
	}
	// It has to name the thing. "Not supported" tells them nothing.
	if !strings.Contains(strings.ToLower(got.RejectionReason), "waste") &&
		!strings.Contains(strings.ToLower(got.RejectionReason), "garbage") {
		t.Errorf("the reason should name what was recognised: %q", got.RejectionReason)
	}
}

func TestACoveredCategoryStillRoutes(t *testing.T) {
	got := Apply(pothole(), CoverageFor("road_defect"))

	if got.Outcome != Accepted {
		t.Errorf("outcome: got %q, want %q", got.Outcome, Accepted)
	}
}

func TestTheNotYetCoveredMessageDoesNotPromiseAnything(t *testing.T) {
	// Hard rule 1 territory: the platform must not imply it will act. "We
	// will look into it" is a promise nobody is keeping.
	r := pothole()
	r.Category = "environment"
	got := Apply(r, CoverageFor("road_defect"))

	lower := strings.ToLower(got.RejectionReason)
	for _, promise := range []string{"we will", "we'll", "will be forwarded", "has been sent"} {
		if strings.Contains(lower, promise) {
			t.Errorf("the message promises action (%q): %q", promise, got.RejectionReason)
		}
	}
}

func TestCoverageDrivesRoutingRatherThanAHardCodedList(t *testing.T) {
	// When waste is added to authority_departments, waste starts routing —
	// without anyone remembering to edit this package.
	r := pothole()
	r.Category = "waste"
	r.Subcategory = "uncollected garbage"

	if got := Apply(r, CoverageFor("road_defect", "waste")); got.Outcome != Accepted {
		t.Errorf("outcome: got %q, want %q once waste is covered", got.Outcome, Accepted)
	}
}
