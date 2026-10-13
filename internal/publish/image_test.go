package publish

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"testing"
)

// street builds something that compresses like a photograph: broad smooth
// regions with fine texture over them, rather than the pixel-by-pixel noise a
// simple formula produces. Noise is incompressible, so a noise fixture would
// hold ForWeb to a standard no real photograph sets.
func street(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		fy := float64(y) / float64(h)
		for x := 0; x < w; x++ {
			fx := float64(x) / float64(w)
			// Smooth tarmac, a brighter kerb across the upper third, and a
			// little grain so it is not a flat fill either.
			base := 60 + 90*math.Abs(math.Sin(3*fx+2*fy))
			if fy < 0.34 {
				base += 70
			}
			grain := float64((x*31 + y*17) % 13)
			v := base + grain
			img.SetNRGBA(x, y, color.NRGBA{
				R: clamp(v + 12), G: clamp(v + 4), B: clamp(v - 6), A: 255,
			})
		}
	}
	return img
}

func clamp(v float64) uint8 {
	switch {
	case v < 0:
		return 0
	case v > 255:
		return 255
	}
	return uint8(v)
}

func encoded(t *testing.T, img image.Image, quality int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestForWebMeetsThePageBudget(t *testing.T) {
	// What a phone sends after the client's own resize, redacted and
	// re-encoded by the redaction pass.
	big := encoded(t, street(2048, 1536), 82)
	if len(big) < Budget {
		t.Fatalf("test input is already under budget at %d bytes; it proves nothing", len(big))
	}

	out, err := ForWeb(big)
	if err != nil {
		t.Fatalf("ForWeb: %v", err)
	}
	if len(out) > Budget {
		t.Errorf("published image is %d bytes, over the %d byte budget a 2G "+
			"connection has for a whole page", len(out), Budget)
	}

	cfg, err := jpeg.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("output is not a readable JPEG: %v", err)
	}
	if cfg.Width > LongSide || cfg.Height > LongSide {
		t.Errorf("published image is %dx%d, want neither side over %d", cfg.Width, cfg.Height, LongSide)
	}
	// Still large enough that the defect is the subject rather than a smudge.
	if cfg.Width < 640 && cfg.Height < 640 {
		t.Errorf("published image is %dx%d, too small to show what was reported", cfg.Width, cfg.Height)
	}
}

func TestForWebKeepsTheShapeOfThePhotograph(t *testing.T) {
	for _, c := range []struct{ w, h int }{{2048, 1536}, {1536, 2048}, {4000, 1000}} {
		out, err := ForWeb(encoded(t, street(c.w, c.h), 82))
		if err != nil {
			t.Fatalf("%dx%d: %v", c.w, c.h, err)
		}
		cfg, err := jpeg.DecodeConfig(bytes.NewReader(out))
		if err != nil {
			t.Fatal(err)
		}
		was := float64(c.w) / float64(c.h)
		now := float64(cfg.Width) / float64(cfg.Height)
		if now < was*0.9 || now > was*1.1 {
			t.Errorf("%dx%d became %dx%d: aspect %0.2f became %0.2f",
				c.w, c.h, cfg.Width, cfg.Height, was, now)
		}
	}
}

func TestForWebLeavesASmallPhotographAlone(t *testing.T) {
	// Re-encoding an image that is already within budget only loses quality.
	small := encoded(t, street(800, 600), 70)
	out, err := ForWeb(small)
	if err != nil {
		t.Fatalf("ForWeb: %v", err)
	}
	if !bytes.Equal(out, small) {
		t.Errorf("an image already within budget was re-encoded: %d bytes became %d", len(small), len(out))
	}
}

func TestForWebRefusesWhatItCannotRead(t *testing.T) {
	// A truncated object must not be served as if it were a photograph: the
	// public page would show a broken image and say nothing about why.
	if _, err := ForWeb([]byte("not a photograph")); err == nil {
		t.Fatal("ForWeb accepted bytes that are not an image")
	}
}
