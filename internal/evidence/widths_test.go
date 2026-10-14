package evidence

import "testing"

// The tables are hand-entered, and a short one silently indexes out of range
// or measures the wrong glyph for everything past the gap.
func TestTheWidthTablesCoverEveryPrintableCharacter(t *testing.T) {
	want := 127 - 32
	if got := len(helvetica); got != want {
		t.Errorf("helvetica has %d widths, want %d", got, want)
	}
	if got := len(helveticaBold); got != want {
		t.Errorf("helvetica-bold has %d widths, want %d", got, want)
	}
	for i, w := range helvetica {
		if w <= 0 || w > 1100 {
			t.Errorf("helvetica[%d] (%q) is %d", i, rune(i+32), w)
		}
	}
	for i, w := range helveticaBold {
		if w <= 0 || w > 1100 {
			t.Errorf("helvetica-bold[%d] (%q) is %d", i, rune(i+32), w)
		}
	}
}

// Bold is wider than regular for letters. If a table were pasted twice this is
// what would catch it.
func TestBoldIsWiderThanRegular(t *testing.T) {
	const sample = "The authority was told on 12 October 2026"
	reg := measure(sample, bodyFace, bodySize)
	bold := measure(sample, boldFace, bodySize)
	if bold <= reg {
		t.Errorf("bold measures %.1f and regular %.1f", bold, reg)
	}
}

// Anything outside ASCII is transliterated before it is drawn, and measuring
// must assume the widest thing it can become or a line can overrun.
func TestNonAsciiIsMeasuredConservatively(t *testing.T) {
	if measure("—", bodyFace, bodySize) != measure("?", bodyFace, bodySize) {
		t.Error("a non-ASCII rune is not measured as the fallback glyph")
	}
}
