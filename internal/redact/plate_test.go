package redact

import (
	"context"
	"image"
	"image/color"
	"math/rand"
	"testing"
)

// noisyWall is a textured background with no plate-like structure in it.
func noisyWall(w, h int, seed int64) *image.NRGBA {
	rng := rand.New(rand.NewSource(seed))
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8(96 + rng.Intn(24))
			img.SetNRGBA(x, y, color.NRGBA{R: v, G: v, B: v, A: 255})
		}
	}
	return img
}

// drawPlate paints a light rectangle with dark character-like bars, which is
// the structure a plate presents to an edge detector: a tight horizontal run
// of strong vertical edges inside a bright box.
func drawPlate(img *image.NRGBA, r image.Rectangle, bars int) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 240, G: 240, B: 232, A: 255})
		}
	}
	step := r.Dx() / (bars + 1)
	barW := max(step/2, 2)
	for i := 1; i <= bars; i++ {
		x0 := r.Min.X + i*step
		for y := r.Min.Y + 4; y < r.Max.Y-4; y++ {
			for x := x0; x < x0+barW && x < r.Max.X; x++ {
				img.SetNRGBA(x, y, color.NRGBA{R: 20, G: 20, B: 20, A: 255})
			}
		}
	}
}

func TestPlatesFindsAPlateLikePatch(t *testing.T) {
	img := noisyWall(400, 300, 1)
	want := image.Rect(120, 180, 260, 215) // 140 x 35, about 4:1
	drawPlate(img, want, 8)

	got, err := NewPlates().Detect(context.Background(), img)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	covered := false
	for _, r := range got {
		if r.Kind != Plate {
			t.Fatalf("region reported as %q by the plate detector", r.Kind)
		}
		if r.Rect.Intersect(want).Dx() >= want.Dx()/2 {
			covered = true
		}
	}
	if !covered {
		t.Fatalf("the plate at %v was not covered by any of %d regions: %v", want, len(got), got)
	}
}

// A plate photographed from two metres away fills far more of the frame than
// one photographed from ten, and both are ordinary captures. The localiser
// closes gaps between character strokes with a constant in pixels, so a plate
// large enough for those gaps to exceed it came back as a row of separate
// strokes and was dropped by the shape filter — the clearest plate in the set
// was the one it could not see.
func TestPlatesFindsAPlateWhateverSizeItIsInTheFrame(t *testing.T) {
	// Same plate, same proportions, four sizes. 4:1, which is the single-row
	// Indian format.
	for _, size := range []struct {
		name string
		w, h int
	}{
		{name: "far away", w: 70, h: 18},
		{name: "across the road", w: 140, h: 35},
		{name: "a few steps away", w: 280, h: 70},
		{name: "filling the frame", w: 560, h: 140},
	} {
		t.Run(size.name, func(t *testing.T) {
			img := noisyWall(1200, 900, 3)
			want := image.Rect(300, 400, 300+size.w, 400+size.h)
			drawPlate(img, want, 8)

			got, err := NewPlates().Detect(context.Background(), img)
			if err != nil {
				t.Fatalf("Detect: %v", err)
			}
			for _, r := range got {
				if r.Rect.Intersect(want).Dx() >= want.Dx()/2 {
					return
				}
			}
			t.Fatalf("a %dx%d plate at %v was not covered by any of %d regions: %v",
				size.w, size.h, want, len(got), got)
		})
	}
}

func TestPlatesFindsNothingOnAPlainWall(t *testing.T) {
	got, err := NewPlates().Detect(context.Background(), noisyWall(400, 300, 2))
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("found %d plates on a blank textured wall: %v", len(got), got)
	}
}

func TestPlausiblePlateRejectsShapesNoPlateHas(t *testing.T) {
	cases := []struct {
		name string
		rect image.Rectangle
		want bool
	}{
		{"single-row Indian plate, about 4.7:1", image.Rect(0, 0, 141, 30), true},
		{"two-row plate, about 2:1", image.Rect(0, 0, 100, 50), true},
		{"a square patch of gravel", image.Rect(0, 0, 40, 40), false},
		{"a lamp post", image.Rect(0, 0, 12, 200), false},
		{"a kerb line running the whole frame", image.Rect(0, 0, 1200, 30), false},
		{"too small to read anyway", image.Rect(0, 0, 18, 5), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := plausiblePlate(c.rect); got != c.want {
				t.Fatalf("plausiblePlate(%v) = %v, want %v", c.rect, got, c.want)
			}
		})
	}
}

func TestPlatesReportsItself(t *testing.T) {
	p := NewPlates()
	if p.Kind() != Plate {
		t.Fatalf("Kind = %q, want %q", p.Kind(), Plate)
	}
	if p.Tool() == "" {
		t.Fatal("Tool is empty; Decide holds on a check that names no tool")
	}
}

func TestPlatesScansAnImageWhoseBoundsDoNotStartAtTheOrigin(t *testing.T) {
	full := noisyWall(500, 400, 3)
	drawPlate(full, image.Rect(220, 220, 350, 252), 8)
	sub := full.SubImage(image.Rect(150, 150, 450, 350))

	got, err := NewPlates().Detect(context.Background(), sub)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("found nothing in a sub-image containing a plate")
	}
	for _, r := range got {
		if !r.Rect.Overlaps(sub.Bounds()) {
			t.Fatalf("region %v lies outside the image it was found in, %v", r.Rect, sub.Bounds())
		}
	}
}

func TestPlatesStopsOnACancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewPlates().Detect(ctx, noisyWall(200, 200, 4)); err == nil {
		t.Fatal("a cancelled scan reported success")
	}
}

func TestPlatesDoesNotFloodTheImageWithRegions(t *testing.T) {
	// A detector biased towards over-detection still has to leave a usable
	// photograph behind. A busy street scene must not come back as a hundred
	// mosaic squares.
	img := noisyWall(800, 600, 5)
	for i := 0; i < 4; i++ {
		drawPlate(img, image.Rect(60+i*180, 400, 60+i*180+140, 435), 8)
	}

	got, err := NewPlates().Detect(context.Background(), img)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(got) > maxPlateRegions {
		t.Fatalf("found %d regions, want at most %d", len(got), maxPlateRegions)
	}
}
