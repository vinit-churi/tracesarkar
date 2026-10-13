// Package redact obscures faces and number plates in a photograph, and decides
// whether what it managed to do is enough to publish the result.
//
// Hard rule 5 says no raw face or number plate reaches a public derivative. The
// package is built so that the rule is satisfied by what happened rather than by
// what was believed: Decide refuses to publish unless every required check can
// account for itself, and Apply destroys the pixels rather than softening them.
//
// Nothing here calls out to a third party. Sending an un-redacted photograph of
// an identifiable person to somebody else's API in order to find the person to
// redact is self-defeating ([ADR 0008]).
//
// [ADR 0008]: ../../docs/04-adr/0008-redaction-approach.md
package redact

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// Kind is a thing redaction looks for.
type Kind string

const (
	Face  Kind = "face"
	Plate Kind = "plate"
)

// required is everything that must have been looked for before a photograph may
// be published. Hard rule 5 names both, so a photograph checked for one and not
// the other is not checked.
var required = []Kind{Face, Plate}

// Region is one rectangle to destroy, as a detector reported it.
type Region struct {
	Kind  Kind
	Rect  image.Rectangle
	Score float64
}

// Check is one detector's account of itself. It exists so that a scan that
// found nothing can be told apart from a scan that did not happen — the two
// look identical from the regions alone, and only one of them is safe.
type Check struct {
	Kind Kind
	Tool string
	Err  error
}

// Scan is the whole result of looking at one photograph.
type Scan struct {
	Checks  []Check
	Regions []Region
}

func (s Scan) check(k Kind) *Check {
	for i := range s.Checks {
		if s.Checks[i].Kind == k {
			return &s.Checks[i]
		}
	}
	return nil
}

func (s Scan) count(k Kind) int {
	n := 0
	for _, r := range s.Regions {
		if r.Kind == k {
			n++
		}
	}
	return n
}

// Verdict is whether the redacted image may be published, and if not, why not.
// The reason is shown to the reporter, so it is written for them.
type Verdict struct {
	Publish bool
	Reason  string
}

// Decide reports whether a redacted photograph may become a public derivative.
//
// It holds rather than publishes on every doubt. ADR 0008 settled the trade:
// a delayed publication is a small cost and an un-redacted face is not, so the
// default on any incomplete, failed or contradicted scan is to withhold the
// image and keep the rest of the issue public.
//
// peoplePresent is the classifier's reading of the same photograph. It is used
// only to contradict the detector, never to excuse it: a classifier that says
// nobody is there cannot authorise skipping the face check.
func Decide(s Scan, peoplePresent bool) Verdict {
	for _, kind := range required {
		c := s.check(kind)
		switch {
		case c == nil:
			return Verdict{Reason: fmt.Sprintf(
				"nothing looked for %ss in this photograph, so it is not known whether there are any",
				kind)}
		case c.Err != nil:
			return Verdict{Reason: fmt.Sprintf(
				"the %s check did not finish: %v", kind, c.Err)}
		case c.Tool == "":
			return Verdict{Reason: fmt.Sprintf(
				"the %s check names no tool, so there is no record of what ran", kind)}
		}
	}

	if peoplePresent && s.count(Face) == 0 {
		return Verdict{Reason: "people were seen in this photograph but no face was found to " +
			"obscure. One of those two readings is wrong, and publishing on the hopeful one is " +
			"the only choice that can expose somebody."}
	}

	return Verdict{Publish: true}
}

// padFraction expands a detector's box before anything is destroyed. Boxes that
// clip a jaw, a hairline or the edge of a plate are common enough that the box
// itself is not a safe boundary, and the margin costs a few pixels of road.
const padFraction = 0.12

// minPad keeps the margin meaningful on a small box, where a fraction rounds
// away to nothing.
const minPad = 2

// blocksAcross is how coarse a redacted region becomes: roughly five blocks on
// its shorter side, whatever its size, so a face 16 px wide and a face 400 px
// wide are both unrecognisable.
const blocksAcross = 5

// minBlock stops the grid becoming fine enough to read on a tiny region.
const minBlock = 3

// Apply destroys the given regions of src and returns the result. src is not
// modified: the archival original is the evidentiary record and nothing here
// may touch it.
//
// Each region is averaged down to blocks rather than blurred. A blur is a
// convolution and leaves the original signal attenuated in the result; an
// average over a block discards everything finer than the block, which is what
// "obscured" has to mean for a face.
func Apply(src image.Image, regions []Region) *image.NRGBA {
	b := src.Bounds()
	out := image.NewNRGBA(b)
	draw.Draw(out, b, src, b.Min, draw.Src)

	for _, r := range regions {
		mosaic(out, pad(r.Rect).Intersect(b))
	}
	return out
}

func pad(r image.Rectangle) image.Rectangle {
	if r.Empty() {
		return r
	}
	p := int(float64(shortSide(r)) * padFraction)
	if p < minPad {
		p = minPad
	}
	return r.Inset(-p)
}

func mosaic(img *image.NRGBA, r image.Rectangle) {
	if r.Empty() {
		return
	}
	block := shortSide(r) / blocksAcross
	if block < minBlock {
		block = minBlock
	}
	for y := r.Min.Y; y < r.Max.Y; y += block {
		for x := r.Min.X; x < r.Max.X; x += block {
			cell := image.Rect(x, y, x+block, y+block).Intersect(r)
			fill(img, cell, mean(img, cell))
		}
	}
}

func mean(img *image.NRGBA, r image.Rectangle) color.NRGBA {
	var sr, sg, sb, sa, n uint32
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := img.NRGBAAt(x, y)
			sr += uint32(c.R)
			sg += uint32(c.G)
			sb += uint32(c.B)
			sa += uint32(c.A)
			n++
		}
	}
	if n == 0 {
		return color.NRGBA{}
	}
	return color.NRGBA{
		R: uint8(sr / n),
		G: uint8(sg / n),
		B: uint8(sb / n),
		A: uint8(sa / n),
	}
}

func fill(img *image.NRGBA, r image.Rectangle, c color.NRGBA) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
}

func shortSide(r image.Rectangle) int {
	if r.Dy() < r.Dx() {
		return r.Dy()
	}
	return r.Dx()
}

// Tools names every detector that accounted for itself, in the order they ran.
//
// It is what goes into the redaction record. A detector improvement is rolled
// out by re-scanning every photograph whose record names the old string, so the
// string has to be specific enough to tell versions apart.
func (s Scan) Tools() string {
	out := ""
	for _, c := range s.Checks {
		if c.Tool == "" {
			continue
		}
		if out != "" {
			out += "+"
		}
		out += c.Tool
	}
	return out
}
