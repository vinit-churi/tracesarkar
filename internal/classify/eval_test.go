package classify

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// fakeClassifier answers from a script, so the harness can be tested without
// spending money or depending on a provider being up.
type fakeClassifier struct {
	byLabel map[string]Result
	// errByLabel fails only the images named in it, which is how a run with
	// some unreadable photographs is distinguished from a run where the
	// provider itself is down.
	errByLabel map[string]error
	err        error
	calls      int
}

func (f *fakeClassifier) Classify(_ context.Context, image []byte, _ string) (Result, error) {
	f.calls++
	if f.err != nil {
		return Result{}, f.err
	}
	if e, ok := f.errByLabel[string(image)]; ok {
		return Result{}, e
	}
	return f.byLabel[string(image)], nil
}

func labelled(label, category, sub string) EvalCase {
	return EvalCase{
		ID: label, Image: []byte(label),
		WantCategory: category, WantSubcategory: sub,
	}
}

func TestTheHarnessScoresCategoryAndSubcategorySeparately(t *testing.T) {
	// Getting the category right routes the complaint to the correct
	// department; getting the subcategory right only changes wording. They
	// are not the same failure and must not share a number.
	f := &fakeClassifier{byLabel: map[string]Result{
		"a": {Category: "road_defect", Subcategory: "pothole", Confidence: 0.9, IsCivicIssue: true, ImageQuality: "good"},
		"b": {Category: "road_defect", Subcategory: "crack", Confidence: 0.9, IsCivicIssue: true, ImageQuality: "good"},
		"c": {Category: "waste", Subcategory: "debris", Confidence: 0.9, IsCivicIssue: true, ImageQuality: "good"},
	}}
	cases := []EvalCase{
		labelled("a", "road_defect", "pothole"), // both right
		labelled("b", "road_defect", "pothole"), // category right, sub wrong
		labelled("c", "road_defect", "pothole"), // both wrong
	}

	report, err := RunEval(context.Background(), f, cases, CoverageFor("road_defect", "waste"))
	if err != nil {
		t.Fatalf("RunEval: %v", err)
	}

	if report.CategoryAccuracy() != 2.0/3.0 {
		t.Errorf("category accuracy: got %v, want 2/3", report.CategoryAccuracy())
	}
	if report.SubcategoryAccuracy() != 1.0/3.0 {
		t.Errorf("subcategory accuracy: got %v, want 1/3", report.SubcategoryAccuracy())
	}
}

func TestTheHarnessRecordsEveryMissSoTheyCanBeRead(t *testing.T) {
	// An accuracy number alone cannot be acted on. The failures are the
	// thing worth looking at.
	f := &fakeClassifier{byLabel: map[string]Result{
		"a": {Category: "waste", Subcategory: "debris", Confidence: 0.9, IsCivicIssue: true, ImageQuality: "good"},
	}}

	report, err := RunEval(context.Background(), f, []EvalCase{labelled("a", "road_defect", "pothole")},
		CoverageFor("road_defect", "waste"))
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Misses) != 1 {
		t.Fatalf("misses: got %d, want 1", len(report.Misses))
	}
	m := report.Misses[0]
	if m.Want != "road_defect" || m.Got != "waste" {
		t.Errorf("the miss should say what was expected and what came back: %+v", m)
	}
}

func TestAHazardMissedIsCountedSeparatelyFromAnyOtherError(t *testing.T) {
	// Calling an open manhole safe is not the same kind of mistake as
	// calling a pothole a crack, and averaging them hides the one that
	// matters.
	f := &fakeClassifier{byLabel: map[string]Result{
		"a": {Category: "road_defect", Subcategory: "pothole", HazardToLife: false,
			Confidence: 0.9, IsCivicIssue: true, ImageQuality: "good"},
	}}
	c := labelled("a", "road_defect", "pothole")
	c.WantHazard = true

	report, err := RunEval(context.Background(), f, []EvalCase{c}, CoverageFor("road_defect"))
	if err != nil {
		t.Fatal(err)
	}

	if report.HazardsMissed != 1 {
		t.Errorf("hazards missed: got %d, want 1", report.HazardsMissed)
	}
}

func TestTheHarnessMeasuresWhatThePlatformWouldActuallyDo(t *testing.T) {
	// The eval must score the decision after the guardrails, not the raw
	// model output — otherwise it measures something the citizen never sees.
	f := &fakeClassifier{byLabel: map[string]Result{
		"a": {Category: "road_defect", Subcategory: "open manhole", HazardToLife: false,
			Confidence: 0.9, IsCivicIssue: true, ImageQuality: "good"},
	}}
	c := labelled("a", "road_defect", "open manhole")
	c.WantHazard = true

	report, err := RunEval(context.Background(), f, []EvalCase{c}, CoverageFor("road_defect"))
	if err != nil {
		t.Fatal(err)
	}

	// The override rescues it, so this is not a missed hazard in practice.
	if report.HazardsMissed != 0 {
		t.Errorf("the known-hazard override should have caught this: %d", report.HazardsMissed)
	}
	if report.Overrides != 1 {
		t.Errorf("the override should be counted: %d", report.Overrides)
	}
}

func TestAProviderFailureStopsTheRunRatherThanScoringZero(t *testing.T) {
	// A run that scores a provider outage as 0% accuracy would send someone
	// chasing a prompt problem that does not exist.
	f := &fakeClassifier{err: errors.New("gateway unreachable")}

	_, err := RunEval(context.Background(), f, []EvalCase{labelled("a", "road_defect", "pothole")},
		CoverageFor("road_defect"))

	if err == nil {
		t.Fatal("a provider failure must stop the run")
	}
	if !strings.Contains(err.Error(), "unreachable") {
		t.Errorf("the cause should survive: %v", err)
	}
}

func TestAnEmptyEvalSetIsRefused(t *testing.T) {
	// "100% accurate over zero photographs" is the most dangerous number a
	// report can carry.
	if _, err := RunEval(context.Background(), &fakeClassifier{}, nil, CoverageFor("road_defect")); err == nil {
		t.Fatal("an empty eval set must be an error, not a perfect score")
	}
}
