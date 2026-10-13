package redact

import (
	"context"
	"errors"
	"image"
	"image/color"
	"testing"
)

func TestBoxCentresTheDetectionOnWhatTheCascadeReported(t *testing.T) {
	cases := []struct {
		name            string
		row, col, scale int
		want            image.Rectangle
	}{
		{"mid-frame", 100, 50, 40, image.Rect(30, 80, 70, 120)},
		{"odd scale rounds outwards", 100, 50, 41, image.Rect(29, 79, 71, 121)},
		{"at the origin", 0, 0, 20, image.Rect(-10, -10, 10, 10)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// pigo reports a face as a centre and a size, not a box. Getting
			// the two axes the wrong way round would blur the road and leave
			// the face, and the error is invisible without a photograph.
			got := box(c.row, c.col, c.scale)
			if got != c.want {
				t.Fatalf("box(row=%d, col=%d, scale=%d) = %v, want %v", c.row, c.col, c.scale, got, c.want)
			}
		})
	}
}

func TestFacesLoadsTheEmbeddedCascade(t *testing.T) {
	if len(faceCascade) == 0 {
		t.Fatal("the cascade is not embedded; the detector would never run in production")
	}
	f, err := NewFaces()
	if err != nil {
		t.Fatalf("NewFaces: %v", err)
	}
	if f.Kind() != Face {
		t.Fatalf("Kind = %q, want %q", f.Kind(), Face)
	}
	if f.Tool() == "" {
		t.Fatal("Tool is empty; Decide holds on a check that names no tool")
	}
}

func TestFacesFindsNoFaceInAFlatImage(t *testing.T) {
	f, err := NewFaces()
	if err != nil {
		t.Fatalf("NewFaces: %v", err)
	}
	flat := image.NewNRGBA(image.Rect(0, 0, 160, 160))
	for y := 0; y < 160; y++ {
		for x := 0; x < 160; x++ {
			flat.SetNRGBA(x, y, color.NRGBA{R: 120, G: 120, B: 120, A: 255})
		}
	}

	got, err := f.Detect(context.Background(), flat)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("found %d faces in a flat grey square", len(got))
	}
}

// fake is a detector that reports whatever the test tells it to.
type fake struct {
	kind    Kind
	tool    string
	regions []Region
	err     error
}

func (f fake) Kind() Kind { return f.kind }
func (f fake) Tool() string {
	return f.tool
}
func (f fake) Detect(context.Context, image.Image) ([]Region, error) { return f.regions, f.err }

func TestScannerRecordsWhatEachDetectorDid(t *testing.T) {
	boom := errors.New("cascade would not load")
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

	s := NewScanner(
		fake{kind: Face, tool: "fake-face", regions: []Region{{Kind: Face, Rect: image.Rect(1, 1, 5, 5)}}},
		fake{kind: Plate, tool: "fake-plate", err: boom},
	).Scan(context.Background(), img)

	if len(s.Checks) != 2 {
		t.Fatalf("got %d checks, want one per detector", len(s.Checks))
	}
	face := s.check(Face)
	if face == nil || face.Tool != "fake-face" || face.Err != nil {
		t.Fatalf("face check = %+v, want it to name its tool and report no error", face)
	}
	plate := s.check(Plate)
	if plate == nil || !errors.Is(plate.Err, boom) {
		t.Fatalf("plate check = %+v, want it to carry the detector's error", plate)
	}
	if len(s.Regions) != 1 {
		t.Fatalf("got %d regions, want the one the face detector found", len(s.Regions))
	}

	// And the scan that records a failure must not be publishable.
	if v := Decide(s, false); v.Publish {
		t.Fatal("a scan carrying a detector failure was published")
	}
}

func TestScannerWithADetectorMissingIsNotPublishable(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

	s := NewScanner(fake{kind: Face, tool: "fake-face"}).Scan(context.Background(), img)

	if v := Decide(s, false); v.Publish {
		t.Fatal("a photograph never checked for number plates was published")
	}
}

func TestScannerStopsWhenTheContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

	s := NewScanner(
		fake{kind: Face, tool: "fake-face"},
		fake{kind: Plate, tool: "fake-plate"},
	).Scan(ctx, img)

	if v := Decide(s, false); v.Publish {
		t.Fatal("a cancelled scan was published; an abandoned check is not a clean one")
	}
}

func TestShrinkKeepsTheWorkingImageSmallEnoughToScanThreeTimes(t *testing.T) {
	cases := []struct {
		name       string
		w, h       int
		wantFactor int
	}{
		{"a phone photograph", 4032, 3024, 4},
		{"already small enough", 800, 600, 1},
		{"exactly at the limit", workingSide, 700, 1},
		{"just over the limit", workingSide + 1, 700, 2},
		{"portrait", 3024, 4032, 4},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := gradient(c.w, c.h)
			got, factor := shrink(src)
			if factor != c.wantFactor {
				t.Fatalf("factor = %d, want %d", factor, c.wantFactor)
			}
			if long := longSide(got.Bounds()); long > workingSide {
				t.Fatalf("working image long side = %d, want at most %d", long, workingSide)
			}
			if factor == 1 && got.Bounds() != src.Bounds() {
				t.Fatalf("an image already small enough was resampled: %v", got.Bounds())
			}
		})
	}
}

func TestRestorePutsABoxBackWhereItCameFrom(t *testing.T) {
	cases := []struct {
		name   string
		in     image.Rectangle
		factor int
		origin image.Point
		want   image.Rectangle
	}{
		{"scaled and offset", image.Rect(10, 10, 20, 20), 2, image.Pt(100, 50), image.Rect(120, 70, 140, 90)},
		{"no scaling, no offset", image.Rect(3, 4, 5, 6), 1, image.Pt(0, 0), image.Rect(3, 4, 5, 6)},
		{"offset only", image.Rect(0, 0, 8, 8), 1, image.Pt(7, 7), image.Rect(7, 7, 15, 15)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// A box found in the shrunk, origin-anchored working copy means
			// nothing until it is back in the coordinates of the photograph
			// that Apply will write into.
			if got := restore(c.in, c.factor, c.origin); got != c.want {
				t.Fatalf("restore(%v, %d, %v) = %v, want %v", c.in, c.factor, c.origin, got, c.want)
			}
		})
	}
}

func TestFacesScansAnImageWhoseBoundsDoNotStartAtTheOrigin(t *testing.T) {
	f, err := NewFaces()
	if err != nil {
		t.Fatalf("NewFaces: %v", err)
	}
	// A sub-image carries its parent's coordinates. Treating those as 0,0
	// would scan the wrong pixels and redact the wrong part of the frame.
	full := gradient(400, 400)
	sub := full.SubImage(image.Rect(100, 100, 300, 300))

	got, err := f.Detect(context.Background(), sub)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	for _, r := range got {
		if !r.Rect.Overlaps(sub.Bounds()) {
			t.Fatalf("region %v lies outside the image it was found in, %v", r.Rect, sub.Bounds())
		}
	}
}
