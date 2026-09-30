package classify

import (
	"strings"
	"testing"
)

func TestAModelThatTrainsOnOurDataCannotSeeACitizenPhotograph(t *testing.T) {
	// Citizen photographs contain faces and number plates before redaction.
	// A provider that retains them for training cannot be asked to forget,
	// so a deletion request could never be honoured (hard rule 5, DPDP).
	// Structured output held true so the training check is what is being
	// exercised, not masked by an earlier refusal.
	trains := ModelPolicy{ID: "some/trains-on-input", NoTraining: "none", ZDR: "none",
		StructuredOutput: true, Vision: true}

	err := trains.AllowedFor(CitizenPhotograph)

	if err == nil {
		t.Fatal("a model that trains on its inputs must be refused for citizen photographs")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "train") {
		t.Errorf("the reason should name the problem: %v", err)
	}
}

func TestTheSameModelIsFineForOurOwnDevelopmentData(t *testing.T) {
	// The maintainer's own test photographs are the maintainer's to give.
	trains := ModelPolicy{ID: "some/trains-on-input", NoTraining: "none", ZDR: "none",
		StructuredOutput: true, Vision: true}

	if err := trains.AllowedFor(OwnTestData); err != nil {
		t.Errorf("own data should be allowed: %v", err)
	}
}

func TestAModelWithoutStructuredOutputIsRefusedForClassification(t *testing.T) {
	// CLAUDE.md: never parse free text into a domain object. Free-text
	// parsing fails silently, which is the worst way for a classifier to fail.
	noSchema := ModelPolicy{ID: "stealth/pixel-canary", NoTraining: "all", ZDR: "all",
		StructuredOutput: false, Vision: true}

	err := noSchema.AllowedFor(OwnTestData)

	if err == nil {
		t.Fatal("a model without structured output must be refused for classification")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "structured") {
		t.Errorf("the reason should name the problem: %v", err)
	}
}

func TestAModelThatCannotSeeIsRefused(t *testing.T) {
	blind := ModelPolicy{ID: "some/text-only", NoTraining: "all", ZDR: "all",
		StructuredOutput: true, Vision: false}

	if err := blind.AllowedFor(OwnTestData); err == nil {
		t.Fatal("classification needs vision")
	}
}

func TestAProtectedModelIsAllowedForCitizenPhotographs(t *testing.T) {
	good := ModelPolicy{ID: "alibaba/qwen3.7-flash", NoTraining: "all", ZDR: "all",
		StructuredOutput: true, Vision: true}

	if err := good.AllowedFor(CitizenPhotograph); err != nil {
		t.Errorf("a protected, schema-capable vision model should be allowed: %v", err)
	}
}

func TestPartialProtectionIsNotProtection(t *testing.T) {
	// "some" means some providers behind the model honour it. For a
	// photograph of a stranger's face, "some" is not a guarantee we can
	// repeat to the person in the photograph.
	partial := ModelPolicy{ID: "alibaba/qwen3-vl-instruct", NoTraining: "some", ZDR: "some",
		StructuredOutput: true, Vision: true}

	if err := partial.AllowedFor(CitizenPhotograph); err == nil {
		t.Fatal("partial no-training must not clear a citizen photograph")
	}
	if err := partial.AllowedFor(OwnTestData); err != nil {
		t.Errorf("partial protection is fine for our own data: %v", err)
	}
}
