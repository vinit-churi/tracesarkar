package classify

import (
	"strings"
	"testing"
)

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

		got := Apply(r)

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

	got := Apply(r)

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

	got := Apply(r)

	if got.Outcome != NeedsRetake {
		t.Errorf("outcome: got %q, want %q", got.Outcome, NeedsRetake)
	}
}

func TestSomethingOutOfScopeIsRejectedSpecifically(t *testing.T) {
	r := pothole()
	r.IsCivicIssue = false
	r.Category = "private_property"

	got := Apply(r)

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

	got := Apply(r)

	if got.Outcome != NeedsConfirmation {
		t.Errorf("outcome: got %q, want %q", got.Outcome, NeedsConfirmation)
	}
	if len(got.Alternatives) == 0 {
		t.Error("the citizen needs options to choose between")
	}
}

func TestHighConfidenceOnAKnownCategoryIsAccepted(t *testing.T) {
	got := Apply(pothole())

	if got.Outcome != Accepted {
		t.Errorf("outcome: got %q, want %q", got.Outcome, Accepted)
	}
}

func TestACategoryOutsideTheTaxonomyIsNotAccepted(t *testing.T) {
	// Structured output guarantees the shape, not the vocabulary. A category
	// we do not have a department for cannot be routed.
	r := pothole()
	r.Category = "alien_invasion"

	got := Apply(r)

	if got.Outcome == Accepted {
		t.Errorf("an unknown category must not be accepted: %+v", got)
	}
}

func TestPeoplePresentForcesRedactionBeforeAnythingPublic(t *testing.T) {
	r := pothole()
	r.PeoplePresent = true

	got := Apply(r)

	if !got.RedactionRequired {
		t.Error("a photograph with people in it cannot produce a public derivative unredacted")
	}
}

func TestTheRejectionNeverBlamesTheCitizen(t *testing.T) {
	r := pothole()
	r.IsCivicIssue = false
	r.Category = "private_property"

	got := Apply(r)

	lower := strings.ToLower(got.RejectionReason)
	for _, word := range []string{"invalid", "wrong", "failed", "error", "you should"} {
		if strings.Contains(lower, word) {
			t.Errorf("the rejection reads as a rebuke (%q): %q", word, got.RejectionReason)
		}
	}
}
