package redact

import (
	"context"
	"image"
	"testing"
)

// canned is a detector that returns a prepared answer per call, in the order
// Score walks the cases. It stands in for pigo so that the scoring itself can
// be tested without a photograph of anybody.
type canned struct {
	kind Kind
	runs [][]Region
	n    int
}

func (c *canned) Kind() Kind { return c.kind }
func (c *canned) Tool() string {
	return string(c.kind) + "-stub"
}
func (c *canned) Detect(_ context.Context, _ image.Image) ([]Region, error) {
	if c.n >= len(c.runs) {
		c.n++
		return nil, nil
	}
	r := c.runs[c.n]
	c.n++
	return r, nil
}

// flat is a photograph of nothing, 400x300. The detectors under test are
// canned, so its contents do not matter; its size does, because ground truth
// is held in fractions of it.
func flat() image.Image {
	return image.NewNRGBA(image.Rect(0, 0, 400, 300))
}

// faceAt is ground truth for a face occupying the given fractions.
func faceAt(minX, minY, maxX, maxY float64) Box {
	return Box{Kind: Face, MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
}

func scanner(runs ...[]Region) (*Scanner, *canned) {
	face := &canned{kind: Face, runs: runs}
	// Plates are always looked for and never found, so the gate has its second
	// required check and nothing in these cases turns on plate detection.
	return NewScanner(face, &canned{kind: Plate}), face
}

func one(regions ...Region) [][]Region { return [][]Region{regions} }

func TestAFaceTheDetectorCoveredCountsAsFound(t *testing.T) {
	// The detection sits exactly on the truth box: 40..120 x 30..90 of 400x300.
	s, _ := scanner(one(Region{Kind: Face, Rect: image.Rect(40, 30, 120, 90)})...)

	r := Score(context.Background(), s, []Case{{
		ID:    "a",
		Image: flat(),
		Truth: []Box{faceAt(0.1, 0.1, 0.3, 0.3)},
	}})

	got, ok := r.Kinds[Face].Recall()
	if !ok {
		t.Fatal("recall came back unmeasured over a case that contains a face")
	}
	if got != 1 {
		t.Errorf("recall = %.2f, want 1", got)
	}
	if n := len(r.Misses); n != 0 {
		t.Errorf("misses = %d, want 0: %+v", n, r.Misses)
	}
}

func TestABoxFarLargerThanTheFaceStillCountsAsCovered(t *testing.T) {
	// Intersection over union here is about 0.1. What matters is that every
	// pixel of the face is gone, which it is.
	s, _ := scanner(one(Region{Kind: Face, Rect: image.Rect(0, 0, 260, 200)})...)

	r := Score(context.Background(), s, []Case{{
		ID:    "a",
		Image: flat(),
		Truth: []Box{faceAt(0.1, 0.1, 0.3, 0.3)},
	}})

	if got, _ := r.Kinds[Face].Recall(); got != 1 {
		t.Errorf("recall = %.2f, want 1 — the face was destroyed, generously", got)
	}
}

func TestHalfACoveredFaceIsAMiss(t *testing.T) {
	// The left half of the face is destroyed and the right half is not. A
	// recognisable eye is not a redacted face.
	s, _ := scanner(one(Region{Kind: Face, Rect: image.Rect(40, 30, 78, 90)})...)

	r := Score(context.Background(), s, []Case{{
		ID:         "a",
		Image:      flat(),
		Truth:      []Box{faceAt(0.1, 0.1, 0.3, 0.3)},
		Conditions: []string{"night"},
	}})

	if got, _ := r.Kinds[Face].Recall(); got != 0 {
		t.Errorf("recall = %.2f, want 0", got)
	}
	if len(r.Misses) != 1 {
		t.Fatalf("misses = %d, want 1", len(r.Misses))
	}
	m := r.Misses[0]
	if m.ID != "a" || m.Kind != Face {
		t.Errorf("miss = %+v, want case a, a face", m)
	}
	// The fraction is reported because a miss at 0.9 and a miss at 0.0 are
	// different problems, and the number is what says which.
	if m.Covered <= 0.2 || m.Covered >= 0.8 {
		t.Errorf("covered = %.2f, want roughly half", m.Covered)
	}
	if len(m.Conditions) != 1 || m.Conditions[0] != "night" {
		t.Errorf("conditions = %v, want the case's own", m.Conditions)
	}
}

func TestAPlateCoveredByAFaceMosaicCountsAsCovered(t *testing.T) {
	// Destroyed pixels are destroyed whatever the detector thought it saw. A
	// plate under a box labelled face is not a leak.
	s, _ := scanner(one(Region{Kind: Face, Rect: image.Rect(40, 30, 120, 90)})...)

	r := Score(context.Background(), s, []Case{{
		ID:    "a",
		Image: flat(),
		Truth: []Box{{Kind: Plate, MinX: 0.1, MinY: 0.1, MaxX: 0.3, MaxY: 0.3}},
	}})

	if got, _ := r.Kinds[Plate].Recall(); got != 1 {
		t.Errorf("plate recall = %.2f, want 1", got)
	}
}

func TestRecallIsUnmeasuredRatherThanPerfectWhenNothingIsThereToFind(t *testing.T) {
	// The whole reason this harness exists. Five captures with no face in them
	// produced no recall figure, and a harness that reports 100% over them
	// would have hidden exactly the thing that needed measuring.
	s, _ := scanner(one()...)

	r := Score(context.Background(), s, []Case{{ID: "a", Image: flat()}})

	if _, ok := r.Kinds[Face].Recall(); ok {
		t.Error("recall was reported over a set containing no face")
	}
	if r.Cases != 1 {
		t.Errorf("cases = %d, want 1", r.Cases)
	}
}

func TestAMosaicOverNothingIsSpuriousNotAMiss(t *testing.T) {
	s, _ := scanner(one(
		Region{Kind: Face, Rect: image.Rect(40, 30, 120, 90)},
		Region{Kind: Face, Rect: image.Rect(300, 200, 340, 240)},
	)...)

	r := Score(context.Background(), s, []Case{{
		ID:    "a",
		Image: flat(),
		Truth: []Box{faceAt(0.1, 0.1, 0.3, 0.3)},
	}})

	k := r.Kinds[Face]
	if k.Spurious != 1 {
		t.Errorf("spurious = %d, want 1", k.Spurious)
	}
	if len(r.Misses) != 0 {
		t.Errorf("misses = %d, want 0 — a box on the pavement loses evidence, not a face", len(r.Misses))
	}
	p, ok := k.Precision()
	if !ok || p != 0.5 {
		t.Errorf("precision = %.2f (%v), want 0.5", p, ok)
	}
}

func TestAnUncoveredFaceThatTheGateHeldIsAMissAndNotAnExposure(t *testing.T) {
	// Nothing found, and the classifier saw people. Decide holds, so the face
	// never reaches anybody. It is still a detector miss and still counted.
	s, _ := scanner(one()...)

	r := Score(context.Background(), s, []Case{{
		ID:            "a",
		Image:         flat(),
		Truth:         []Box{faceAt(0.1, 0.1, 0.3, 0.3)},
		PeoplePresent: true,
	}})

	if len(r.Misses) != 1 {
		t.Fatalf("misses = %d, want 1", len(r.Misses))
	}
	if r.Misses[0].Published {
		t.Error("the miss was marked published, but the gate held the image")
	}
	if r.Exposed() != 0 {
		t.Errorf("exposed = %d, want 0", r.Exposed())
	}
	if r.Held != 1 {
		t.Errorf("held = %d, want 1", r.Held)
	}
}

func TestAnUncoveredFaceThePlatformWouldHavePublishedIsAnExposure(t *testing.T) {
	// The dangerous bucket, and the one number this harness exists to produce:
	// a face in the photograph, no face found, and nothing to contradict it.
	s, _ := scanner(one()...)

	r := Score(context.Background(), s, []Case{{
		ID:    "a",
		Image: flat(),
		Truth: []Box{faceAt(0.1, 0.1, 0.3, 0.3)},
		// The classifier did not see people, so Decide has no reason to hold.
		PeoplePresent: false,
	}})

	if r.Exposed() != 1 {
		t.Fatalf("exposed = %d, want 1", r.Exposed())
	}
	if !r.Misses[0].Published {
		t.Error("the miss was not marked published, but the gate let the image out")
	}
}

func TestRecallIsBrokenDownByCondition(t *testing.T) {
	// A detector that works at noon and fails at night has a lighting problem,
	// not an accuracy problem, and the aggregate hides which.
	covered := []Region{{Kind: Face, Rect: image.Rect(40, 30, 120, 90)}}
	s, _ := scanner(covered, nil)

	r := Score(context.Background(), s, []Case{
		{ID: "day", Image: flat(), Truth: []Box{faceAt(0.1, 0.1, 0.3, 0.3)}, Conditions: []string{"daylight"}},
		{ID: "night", Image: flat(), Truth: []Box{faceAt(0.1, 0.1, 0.3, 0.3)}, Conditions: []string{"night"}},
	})

	if got, _ := r.ByCondition["daylight"].Recall(); got != 1 {
		t.Errorf("daylight recall = %.2f, want 1", got)
	}
	if got, _ := r.ByCondition["night"].Recall(); got != 0 {
		t.Errorf("night recall = %.2f, want 0", got)
	}
	if got, _ := r.Kinds[Face].Recall(); got != 0.5 {
		t.Errorf("overall recall = %.2f, want 0.5", got)
	}
}

func TestTheReportNamesTheToolsItScored(t *testing.T) {
	// A recall figure with no tool string against it cannot be re-checked
	// after the detector changes.
	s, _ := scanner(one()...)

	r := Score(context.Background(), s, []Case{{ID: "a", Image: flat()}})

	if r.Tools != "face-stub+plate-stub" {
		t.Errorf("tools = %q, want the detectors that ran", r.Tools)
	}
}

func TestABoxTurnsIntoPixelsOfWhateverImageItIsAskedAbout(t *testing.T) {
	b := faceAt(0.1, 0.2, 0.5, 0.6)

	got := b.Rect(image.Rect(0, 0, 400, 300))
	if want := image.Rect(40, 60, 200, 180); got != want {
		t.Errorf("rect = %v, want %v", got, want)
	}

	// The same truth against a thumbnail of the same photograph.
	got = b.Rect(image.Rect(0, 0, 200, 150))
	if want := image.Rect(20, 30, 100, 90); got != want {
		t.Errorf("rect on a half-size copy = %v, want %v", got, want)
	}
}

func TestABoxOutsideThePhotographIsRefused(t *testing.T) {
	cases := []struct {
		name string
		box  Box
		ok   bool
	}{
		{name: "a face", box: faceAt(0.1, 0.1, 0.3, 0.3), ok: true},
		{name: "the whole photograph", box: faceAt(0, 0, 1, 1), ok: true},
		{name: "a plate", box: Box{Kind: Plate, MaxX: 0.5, MaxY: 0.5}, ok: true},
		{name: "past the right edge", box: faceAt(0.5, 0.1, 1.2, 0.3)},
		{name: "starting off the top", box: faceAt(0.1, -0.1, 0.3, 0.3)},
		{name: "inside out", box: faceAt(0.5, 0.1, 0.2, 0.3)},
		{name: "a point, not a box", box: faceAt(0.5, 0.5, 0.5, 0.5)},
		{name: "a kind redaction does not look for", box: Box{Kind: "tattoo", MaxX: 0.5, MaxY: 0.5}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.box.Valid()
			if c.ok && err != nil {
				t.Errorf("Valid() = %v, want accepted", err)
			}
			if !c.ok && err == nil {
				t.Error("Valid() accepted a box that cannot be scored")
			}
		})
	}
}

func TestOnlyConditionsTheBreakdownKnowsAreAccepted(t *testing.T) {
	// A condition typed by hand splits the breakdown it was meant to inform:
	// "nite" and "night" are two rows of one case each, and neither says
	// anything. The vocabulary is fixed for the same reason the taxonomy is.
	for _, c := range Conditions() {
		if !KnownCondition(c) {
			t.Errorf("%q is offered to the annotator and then refused", c)
		}
	}
	for _, c := range []string{"nite", "", "NIGHT", "something else"} {
		if KnownCondition(c) {
			t.Errorf("%q was accepted into the breakdown", c)
		}
	}
	// The conditions ADR 0008 names specifically must be among them, or the
	// set cannot answer the question the ADR asked.
	for _, needed := range []string{"night", "distance", "partial-occlusion", "reflection", "plate-at-angle"} {
		if !KnownCondition(needed) {
			t.Errorf("ADR 0008 requires %q to be measurable and it is not offered", needed)
		}
	}
}
