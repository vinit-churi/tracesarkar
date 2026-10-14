package redact

import (
	"context"
	"fmt"
	"image"
)

// This file scores the detectors against photographs a person has annotated:
// the adversarial set ADR 0008 has required from the start. It exists because
// the detector's recall was unknown — the captures it had been run over
// contained no face and no plate, so there was nothing it could have missed,
// and no number anybody could act on.
//
// Two decisions shape everything below.
//
// Recall is measured as destroyed pixels, not as box overlap. Intersection
// over union is the usual metric and it is the wrong one here: a detection
// three times the size of a face scores badly and yet obliterates it, while a
// detection that fits the face neatly but stops at the cheekbones scores well
// and leaves an eye. The question the hard rule asks is what survived.
//
// A figure is withheld rather than defaulted. Recall over a set with nothing
// to find is not 100%, it is unmeasured, and the harness says so.

// coveredAt is the fraction of a face or plate that must be destroyed before it
// counts as covered. It is high on purpose: the remainder is the part somebody
// could still recognise.
const coveredAt = 0.95

// Box is ground truth for one face or plate, in fractions of the photograph's
// width and height.
//
// Fractions rather than pixels so that the annotation survives the image being
// resized — it is marked up on a copy small enough to send to a browser and
// scored against the original the detector actually reads.
type Box struct {
	Kind Kind    `json:"kind"`
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
}

// Rect places the box on a particular copy of the photograph.
func (b Box) Rect(bounds image.Rectangle) image.Rectangle {
	w, h := float64(bounds.Dx()), float64(bounds.Dy())
	return image.Rect(
		bounds.Min.X+int(b.MinX*w+0.5),
		bounds.Min.Y+int(b.MinY*h+0.5),
		bounds.Min.X+int(b.MaxX*w+0.5),
		bounds.Min.Y+int(b.MaxY*h+0.5),
	)
}

// Valid reports whether a box can be scored at all. An annotation that reaches
// past the edge of the photograph, or has no area, or names something the
// detectors do not look for, would quietly distort every figure computed from
// the set, so it is refused where it is entered rather than averaged in.
func (b Box) Valid() error {
	known := false
	for _, k := range required {
		if b.Kind == k {
			known = true
		}
	}
	if !known {
		return fmt.Errorf("redaction looks for %v, not %q", required, b.Kind)
	}
	for _, v := range []float64{b.MinX, b.MinY, b.MaxX, b.MaxY} {
		if v < 0 || v > 1 {
			return fmt.Errorf("a box is marked in fractions of the photograph, and %v is outside it", v)
		}
	}
	if b.MaxX <= b.MinX || b.MaxY <= b.MinY {
		return fmt.Errorf("a box with no area cannot be covered or missed")
	}
	return nil
}

// conditions is the fixed vocabulary the breakdown is grouped by. It is the
// list ADR 0008 asked for, plus what fieldwork in Mumbai actually produces.
//
// Fixed rather than free text for the same reason the classification taxonomy
// is: a condition typed by hand splits its own row, and a breakdown of
// one-case rows says nothing about anything.
var conditions = []string{
	// What the ADR named.
	"distance",
	"partial-occlusion",
	"reflection",
	"night",
	"plate-at-angle",
	// Light and weather, which is what usually explains a miss.
	"daylight",
	"low-light",
	"backlit",
	"rain",
	// How the face sits in the frame. The cascade is frontal; a profile is a
	// different question and has to be countable separately.
	"side-profile",
	"face-turned-away",
	"crowd",
	"motion-blur",
	// Indian plate formats specifically, which the ADR singled out.
	"plate-two-wheeler",
	"plate-commercial-yellow",
	"plate-handpainted",
	"plate-partly-obscured",
}

// Conditions is what an annotator may choose from.
func Conditions() []string {
	out := make([]string, len(conditions))
	copy(out, conditions)
	return out
}

// KnownCondition reports whether a condition belongs to the vocabulary.
func KnownCondition(c string) bool {
	for _, k := range conditions {
		if k == c {
			return true
		}
	}
	return false
}

// Case is one annotated photograph of the set.
type Case struct {
	ID    string
	Image image.Image
	// Truth is every face and plate a person could see in the photograph. An
	// empty list is a claim too: that there is nothing here to find.
	Truth []Box
	// Conditions are the labeller's words for what makes the photograph hard —
	// night, rain, distance, angle, reflection. Recall is broken down by them
	// because the aggregate hides which conditions the detector cannot handle.
	Conditions []string
	// PeoplePresent is what a person can see rather than what the classifier
	// said, so the contradiction rule in Decide is exercised as it would be by
	// a classifier that reads the photograph correctly.
	PeoplePresent bool
}

// Tally is what was found and what was missed, for one kind or one condition.
type Tally struct {
	// Truth is how many faces or plates were there to find.
	Truth int
	// Covered is how many of them the applied regions destroyed.
	Covered int
	// Found is how many regions the detector reported.
	Found int
	// Spurious is how many of those sat on nothing annotated. They are not a
	// privacy failure; they destroy evidence, which is the other cost.
	Spurious int
}

// Recall reports the share of real faces or plates that were destroyed, and
// whether the figure means anything. It means nothing when there was nothing
// to find, and a harness that returned 1.0 in that case would have reported a
// perfect score for the exact gap this set was built to close.
func (t Tally) Recall() (float64, bool) {
	if t.Truth == 0 {
		return 0, false
	}
	return float64(t.Covered) / float64(t.Truth), true
}

// Precision reports the share of reported regions that sat on something real.
func (t Tally) Precision() (float64, bool) {
	if t.Found == 0 {
		return 0, false
	}
	return float64(t.Found-t.Spurious) / float64(t.Found), true
}

// Miss is one face or plate that came through the redaction still readable.
type Miss struct {
	ID   string
	Kind Kind
	Box  Box
	// Covered is the fraction of it that was destroyed. A miss at 0.9 and a
	// miss at 0.0 are different problems and the number says which.
	Covered float64
	// Published is whether the gate would have let this image out anyway. A
	// miss the gate held is a detector fault; a miss it published is exposure.
	Published  bool
	Conditions []string
}

// Report is the whole scoring run.
type Report struct {
	Cases int
	// Tools names the detectors that produced it, so the figure can be tied to
	// a version and re-checked when one changes.
	Tools       string
	Kinds       map[Kind]Tally
	ByCondition map[string]Tally
	Misses      []Miss
	// Held is how many of the photographs the gate withheld, for any reason.
	Held int
}

// Exposed counts the faces and plates that would have reached a public image.
// It is the number to read first, and it is a count rather than a percentage:
// "99% recall" reads as a pass, "one face published" reads as what it is.
func (r Report) Exposed() int {
	n := 0
	for _, m := range r.Misses {
		if m.Published {
			n++
		}
	}
	return n
}

// Score runs the detectors over every case and measures what they destroyed.
func Score(ctx context.Context, s *Scanner, cases []Case) Report {
	r := Report{
		Cases:       len(cases),
		Kinds:       map[Kind]Tally{},
		ByCondition: map[string]Tally{},
	}
	for _, k := range required {
		r.Kinds[k] = Tally{}
	}

	for _, c := range cases {
		scan := s.Scan(ctx, c.Image)
		r.Tools = scan.Tools()
		verdict := Decide(scan, c.PeoplePresent)
		if !verdict.Publish {
			r.Held++
		}

		bounds := c.Image.Bounds()
		destroyed := applied(scan.Regions, bounds)

		for _, b := range c.Truth {
			kind := r.Kinds[b.Kind]
			kind.Truth++

			share := coverage(b.Rect(bounds), destroyed)
			if share >= coveredAt {
				kind.Covered++
			} else {
				r.Misses = append(r.Misses, Miss{
					ID:         c.ID,
					Kind:       b.Kind,
					Box:        b,
					Covered:    share,
					Published:  verdict.Publish,
					Conditions: c.Conditions,
				})
			}
			r.Kinds[b.Kind] = kind

			for _, cond := range c.Conditions {
				t := r.ByCondition[cond]
				t.Truth++
				if share >= coveredAt {
					t.Covered++
				}
				r.ByCondition[cond] = t
			}
		}

		for i, region := range scan.Regions {
			kind := r.Kinds[region.Kind]
			kind.Found++
			if !touchesAnything(destroyed[i], c.Truth, bounds) {
				kind.Spurious++
			}
			r.Kinds[region.Kind] = kind

			for _, cond := range c.Conditions {
				t := r.ByCondition[cond]
				t.Found++
				if !touchesAnything(destroyed[i], c.Truth, bounds) {
					t.Spurious++
				}
				r.ByCondition[cond] = t
			}
		}
	}
	return r
}

// applied is where the pixels actually go, region for region. It pads exactly
// as Apply pads, because scoring anything other than what the publisher writes
// would measure a pipeline nobody runs.
func applied(regions []Region, bounds image.Rectangle) []image.Rectangle {
	out := make([]image.Rectangle, len(regions))
	for i, r := range regions {
		out[i] = pad(r.Rect).Intersect(bounds)
	}
	return out
}

// coverage is the fraction of want that the destroyed rectangles between them
// cover. It counts pixels rather than adding areas, so rectangles that overlap
// each other are not counted twice.
func coverage(want image.Rectangle, destroyed []image.Rectangle) float64 {
	if want.Empty() {
		return 1
	}
	hit := 0
	for y := want.Min.Y; y < want.Max.Y; y++ {
		for x := want.Min.X; x < want.Max.X; x++ {
			for _, d := range destroyed {
				if (image.Point{X: x, Y: y}).In(d) {
					hit++
					break
				}
			}
		}
	}
	return float64(hit) / float64(want.Dx()*want.Dy())
}

// touchesAnything reports whether a region overlaps any annotated face or
// plate, of any kind. Kind is deliberately ignored: a box labelled face that
// lands on a number plate is doing the right thing by accident, not a mistake.
func touchesAnything(region image.Rectangle, truth []Box, bounds image.Rectangle) bool {
	for _, b := range truth {
		if region.Overlaps(b.Rect(bounds)) {
			return true
		}
	}
	return false
}
