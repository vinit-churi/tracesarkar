package redact

import (
	"errors"
	"image"
	"image/color"
	"testing"
)

// A scan where both required checks ran cleanly.
func clean(regions ...Region) Scan {
	return Scan{
		Checks: []Check{
			{Kind: Face, Tool: "pigo/1.4.6"},
			{Kind: Plate, Tool: "edge-density/1"},
		},
		Regions: regions,
	}
}

func TestDecideHoldsUnlessEverythingWasChecked(t *testing.T) {
	boom := errors.New("cascade would not load")

	cases := []struct {
		name          string
		scan          Scan
		peoplePresent bool
		publish       bool
		// A fragment the reason must contain, so a held image can say why.
		says string
	}{
		{
			name:    "both checks ran, nothing found",
			scan:    clean(),
			publish: true,
		},
		{
			name:    "both checks ran, regions found",
			scan:    clean(Region{Kind: Face, Rect: image.Rect(10, 10, 40, 40)}),
			publish: true,
		},
		{
			name: "plates were never looked for",
			scan: Scan{Checks: []Check{{Kind: Face, Tool: "pigo/1.4.6"}}},
			says: "plate",
		},
		{
			name: "faces were never looked for",
			scan: Scan{Checks: []Check{{Kind: Plate, Tool: "edge-density/1"}}},
			says: "face",
		},
		{
			name: "nothing ran at all",
			scan: Scan{},
			says: "face",
		},
		{
			name: "a check ran and failed",
			scan: Scan{Checks: []Check{
				{Kind: Face, Tool: "pigo/1.4.6", Err: boom},
				{Kind: Plate, Tool: "edge-density/1"},
			}},
			says: "cascade would not load",
		},
		{
			name: "a check claims to have run but names no tool",
			scan: Scan{Checks: []Check{
				{Kind: Face},
				{Kind: Plate, Tool: "edge-density/1"},
			}},
			says: "face",
		},
		{
			// The classifier saw somebody. The detector did not. One of them is
			// wrong, and publishing on the optimistic reading is the one choice
			// that can expose a face.
			name:          "the classifier contradicts the face detector",
			scan:          clean(Region{Kind: Plate, Rect: image.Rect(0, 0, 10, 5)}),
			peoplePresent: true,
			says:          "no face was found",
		},
		{
			name:          "the classifier agrees with the face detector",
			scan:          clean(Region{Kind: Face, Rect: image.Rect(10, 10, 40, 40)}),
			peoplePresent: true,
			publish:       true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Decide(c.scan, c.peoplePresent)
			if got.Publish != c.publish {
				t.Fatalf("Publish = %v, want %v (reason: %q)", got.Publish, c.publish, got.Reason)
			}
			if c.publish {
				return
			}
			if got.Reason == "" {
				t.Fatal("held with no reason; a held image must be able to say why")
			}
			if c.says != "" && !contains(got.Reason, c.says) {
				t.Fatalf("reason = %q, want it to mention %q", got.Reason, c.says)
			}
		})
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// gradient builds an image where every pixel differs from its neighbours, so
// any averaging is visible in the result.
func gradient(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8((x * 7) % 256),
				G: uint8((y * 11) % 256),
				B: uint8((x * y) % 256),
				A: 255,
			})
		}
	}
	return img
}

func TestApplyDestroysDetailInsideTheRegion(t *testing.T) {
	src := gradient(200, 200)
	r := image.Rect(60, 60, 140, 140)

	out := Apply(src, []Region{{Kind: Face, Rect: r}})

	// Detail is measured as the number of distinct colours the region holds.
	// Softening a face leaves hundreds; destroying it leaves a handful.
	before := distinct(src, r)
	after := distinct(out, r)
	if before < 100 {
		t.Fatalf("test image is too flat to prove anything: %d distinct colours", before)
	}
	if after > maxBlocks {
		t.Fatalf("region still holds %d distinct colours, want at most %d", after, maxBlocks)
	}
	if after*4 > before {
		t.Fatalf("region went from %d distinct colours to %d; that is softening, not redaction", before, after)
	}

	// And the region must not simply have been copied through.
	if out.NRGBAAt(100, 100) == src.NRGBAAt(100, 100) {
		t.Fatal("pixel inside the region is unchanged")
	}
}

// maxBlocks is the coarsest grid a redacted region may still show. Six across
// is already unrecognisable; the assertion only has to rule out a soft blur.
const maxBlocks = 36

func distinct(img *image.NRGBA, r image.Rectangle) int {
	seen := map[color.NRGBA]struct{}{}
	r = r.Intersect(img.Bounds())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			seen[img.NRGBAAt(x, y)] = struct{}{}
		}
	}
	return len(seen)
}

func TestApplyLeavesTheRestOfThePhotographAlone(t *testing.T) {
	src := gradient(200, 200)
	// A region in the corner, padded generously by Apply; sample far away.
	out := Apply(src, []Region{{Kind: Face, Rect: image.Rect(0, 0, 30, 30)}})

	for _, p := range []image.Point{{X: 150, Y: 150}, {X: 199, Y: 199}, {X: 120, Y: 10}} {
		if got, want := out.NRGBAAt(p.X, p.Y), src.NRGBAAt(p.X, p.Y); got != want {
			t.Fatalf("pixel at %v changed: %v, want %v — redaction must not destroy the defect", p, got, want)
		}
	}
}

func TestApplyPadsBeyondWhatTheDetectorClaimed(t *testing.T) {
	src := gradient(200, 200)
	r := image.Rect(80, 80, 120, 120)

	out := Apply(src, []Region{{Kind: Face, Rect: r}})

	// A detector box clips the jaw and hairline often enough that the box
	// itself is not a safe boundary. Just outside it must also be gone.
	just := image.Pt(r.Min.X-2, r.Min.Y+10)
	if out.NRGBAAt(just.X, just.Y) == src.NRGBAAt(just.X, just.Y) {
		t.Fatalf("pixel at %v, just outside the detector's box, is unchanged", just)
	}
}

func TestApplyHandlesRegionsTheDetectorPutOffTheEdge(t *testing.T) {
	src := gradient(60, 60)

	cases := []struct {
		name string
		rect image.Rectangle
	}{
		{"over the top-left corner", image.Rect(-20, -20, 15, 15)},
		{"over the bottom-right corner", image.Rect(50, 50, 90, 90)},
		{"wholly outside", image.Rect(300, 300, 340, 340)},
		{"empty", image.Rect(10, 10, 10, 10)},
		{"inverted", image.Rect(40, 40, 10, 10)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Apply(src, []Region{{Kind: Face, Rect: c.rect}})
			if out.Bounds() != src.Bounds() {
				t.Fatalf("bounds changed: %v, want %v", out.Bounds(), src.Bounds())
			}
		})
	}
}

func TestApplyRedactsASmallRegionFinelyEnoughToMatter(t *testing.T) {
	// A face 16 px across must not come back as one flat square either; the
	// point is that no detail survives, at any size.
	src := gradient(100, 100)
	r := image.Rect(40, 40, 56, 56)
	out := Apply(src, []Region{{Kind: Face, Rect: r}})

	before, after := distinct(src, r), distinct(out, r)
	if after >= before {
		t.Fatalf("small region went from %d distinct colours to %d", before, after)
	}
	if after > maxBlocks {
		t.Fatalf("small region still holds %d distinct colours, want at most %d", after, maxBlocks)
	}
}

func TestApplyDoesNotMutateTheOriginal(t *testing.T) {
	src := gradient(80, 80)
	before := src.NRGBAAt(40, 40)

	Apply(src, []Region{{Kind: Face, Rect: image.Rect(20, 20, 60, 60)}})

	if src.NRGBAAt(40, 40) != before {
		t.Fatal("Apply wrote into the source; the archival original must survive untouched")
	}
}

func TestScanNamesEveryToolThatRan(t *testing.T) {
	s := Scan{Checks: []Check{
		{Kind: Face, Tool: "pigo-1.4.6/facefinder"},
		{Kind: Plate, Tool: "edge-density-1"},
	}}
	// The record has to be specific enough to re-run: when a detector
	// improves, the corpus is re-scanned by comparing this string.
	want := "pigo-1.4.6/facefinder+edge-density-1"
	if got := s.Tools(); got != want {
		t.Fatalf("Tools() = %q, want %q", got, want)
	}
}

func TestScanNamesNoToolWhenNothingRan(t *testing.T) {
	if got := (Scan{}).Tools(); got != "" {
		t.Fatalf("Tools() = %q on an empty scan, want empty so Decide holds", got)
	}
}
