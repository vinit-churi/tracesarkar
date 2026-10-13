package publish

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
)

// The budget a public issue page has to meet.
//
// A page that takes a minute to arrive on a 2G connection is a page that most
// of the people this platform is for will never see. The whole page is costed
// at 150 KB, and the photograph is nearly all of it.
const (
	Budget   = 150 * 1024
	LongSide = 1280

	// quality is the starting point. ForWeb steps down from here rather than
	// guessing one number that fits every photograph: a pile of waste
	// compresses nothing like an empty stretch of road.
	quality     = 78
	minQuality  = 45
	qualityStep = 11
)

// ForWeb resizes and re-encodes a redacted photograph to meet the budget.
//
// It operates on the redacted copy, never the original: by the time anything
// gets here the faces are already gone, and shrinking an image is not a
// privacy measure — a face small enough to be unrecognisable to a reader is
// still recoverable by anyone who upsamples it.
//
// An image already within budget is returned exactly as it arrived. Re-encoding
// it would cost quality and save nothing.
func ForWeb(raw []byte) ([]byte, error) {
	if len(raw) <= Budget {
		if _, _, err := image.DecodeConfig(bytes.NewReader(raw)); err != nil {
			return nil, fmt.Errorf("publish: not a readable image: %w", err)
		}
		return raw, nil
	}

	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("publish: not a readable image: %w", err)
	}

	fitted := fit(src, LongSide)
	for q := quality; ; q -= qualityStep {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, fitted, &jpeg.Options{Quality: q}); err != nil {
			return nil, fmt.Errorf("publish: encode: %w", err)
		}
		if buf.Len() <= Budget || q <= minQuality {
			// Below minQuality the artefacts start to look like the defect,
			// which is worse than a slow page: the point of the photograph is
			// that somebody can see what was reported.
			return buf.Bytes(), nil
		}
	}
}

// fit reduces an image so that neither side exceeds longSide, by averaging
// whole blocks. The factor is an integer, which keeps the arithmetic exact and
// the aspect ratio within a pixel of the original.
func fit(src image.Image, longSide int) *image.NRGBA {
	b := src.Bounds()
	long := b.Dx()
	if b.Dy() > long {
		long = b.Dy()
	}

	factor := 1
	if long > longSide {
		factor = (long + longSide - 1) / longSide
	}

	w, h := max(b.Dx()/factor, 1), max(b.Dy()/factor, 1)
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.SetNRGBA(x, y, mean(src, image.Rect(
				b.Min.X+x*factor, b.Min.Y+y*factor,
				b.Min.X+(x+1)*factor, b.Min.Y+(y+1)*factor,
			).Intersect(b)))
		}
	}
	return out
}

func mean(src image.Image, r image.Rectangle) color.NRGBA {
	var sr, sg, sb, n uint32
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			cr, cg, cb, _ := src.At(x, y).RGBA()
			sr += cr >> 8
			sg += cg >> 8
			sb += cb >> 8
			n++
		}
	}
	if n == 0 {
		return color.NRGBA{A: 255}
	}
	return color.NRGBA{R: uint8(sr / n), G: uint8(sg / n), B: uint8(sb / n), A: 255}
}
