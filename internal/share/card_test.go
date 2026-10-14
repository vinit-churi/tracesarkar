package share

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"testing"
)

// The card is the artefact that travels furthest, because most sharing in
// India is image forwarding. Everything on it has to hold up with no link
// attached.
func TestTheCardCarriesTheMarkAndThePermalink(t *testing.T) {
	c := CardFor(lodgedFacts())
	all := strings.Join(c.Lines(), "\n")
	if !strings.Contains(all, "tracesarkar.org/issue/8f2a1c") {
		t.Errorf("the card carries no permalink — a forwarded image without "+
			"provenance is a rumour:\n%s", all)
	}
	if !strings.Contains(all, "TraceSarkar") {
		t.Errorf("the card carries no mark:\n%s", all)
	}
}

// Hard rule 4. A card is forwarded to strangers; an exact coordinate on it
// identifies where somebody stood.
func TestTheCardNeverCarriesAnExactCoordinate(t *testing.T) {
	c := CardFor(lodgedFacts())
	for _, l := range c.Lines() {
		// Six decimal places is a doorstep. Anything on this card that looks
		// like a precise fix is a bug, so the card carries no coordinate at
		// all and says where the position can be seen instead.
		if strings.Count(l, ".") > 0 && strings.ContainsAny(l, "0123456789") {
			if looksLikeCoordinate(l) {
				t.Errorf("the card carries something shaped like a coordinate: %q", l)
			}
		}
	}
}

// Hard rule 3, on the artefact hardest to correct once it is forwarded.
func TestTheCardAssertsNoWrongdoing(t *testing.T) {
	for _, st := range []string{"open", "claimed_resolved", "citizen_confirmed", "reopened"} {
		f := lodgedFacts()
		f.Status = st
		c := CardFor(f)
		if vs := Check(strings.Join(c.Lines(), "\n")); len(vs) != 0 {
			t.Errorf("%s: the card fails the tone check %+v", st, vs)
		}
	}
}

// Hard rule 2 applied to the card: a contractor is named only when the match
// is published, and today nothing is, so the line must be absent rather than
// blank or guessed.
func TestTheCardOmitsAContractLineItDoesNotHave(t *testing.T) {
	all := strings.Join(CardFor(lodgedFacts()).Lines(), "\n")
	for _, word := range []string{"Contract", "Contractor", "Warranty"} {
		if strings.Contains(all, word) {
			t.Errorf("the card printed a %q line with nothing behind it:\n%s", word, all)
		}
	}

	f := lodgedFacts()
	f.Contract = &Contract{
		Code: "WS/2023/ROAD/117", SourceRef: "https://example.test/tender",
	}
	all = strings.Join(CardFor(f).Lines(), "\n")
	if !strings.Contains(all, "WS/2023/ROAD/117") {
		t.Errorf("a published contract did not reach the card:\n%s", all)
	}
}

// claimed_resolved and citizen_confirmed must never share a label (G14), on
// the card as much as anywhere else.
func TestTheCardDistinguishesWhoSaidItWasFixed(t *testing.T) {
	claimed, citizen := lodgedFacts(), lodgedFacts()
	claimed.Status, citizen.Status = "claimed_resolved", "citizen_confirmed"
	a := strings.Join(CardFor(claimed).Lines(), "\n")
	b := strings.Join(CardFor(citizen).Lines(), "\n")
	if a == b {
		t.Error("the card reads the same whether the authority or a resident said so")
	}
}

// Draw has to produce a real image from a real photograph, taller than the
// photograph it was given, because the fact block is added below it rather
// than printed over the evidence.
func TestDrawAddsTheFactBlockBelowThePhotographWithoutCoveringIt(t *testing.T) {
	photo := image.NewRGBA(image.Rect(0, 0, 800, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			photo.Set(x, y, color.RGBA{R: 200, G: 40, B: 40, A: 255})
		}
	}
	out, err := Draw(photo, CardFor(lodgedFacts()))
	if err != nil {
		t.Fatalf("Draw: %v", err)
	}
	b := out.Bounds()
	if b.Dx() != 800 {
		t.Errorf("width changed: %d, want 800", b.Dx())
	}
	if b.Dy() <= 600 {
		t.Errorf("height is %d; the fact block must be added below the "+
			"photograph, not drawn over it", b.Dy())
	}
	// The photograph itself must survive untouched.
	if got := out.At(400, 300); got != (color.RGBA{R: 200, G: 40, B: 40, A: 255}) {
		r, g, bb, _ := got.RGBA()
		t.Errorf("the photograph was altered: %d,%d,%d", r>>8, g>>8, bb>>8)
	}
}

// The card is produced from the redacted copy and encoded for the web, inside
// the same budget the public page is costed at.
func TestEncodeProducesAJpegInsideTheTransferBudget(t *testing.T) {
	photo := image.NewRGBA(image.Rect(0, 0, 1600, 1200))
	for y := 0; y < 1200; y++ {
		for x := 0; x < 1600; x++ {
			photo.Set(x, y, color.RGBA{R: uint8(x % 251), G: uint8(y % 241), B: 90, A: 255})
		}
	}
	body, err := Encode(photo, CardFor(lodgedFacts()))
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(body)); err != nil {
		t.Fatalf("Encode did not produce a decodable jpeg: %v", err)
	}
	if len(body) > CardBudgetBytes {
		t.Errorf("card is %d bytes, budget is %d", len(body), CardBudgetBytes)
	}
}

// Two people generating the card for the same issue must get the same bytes,
// or the artefact cannot be cited.
func TestTheCardIsDeterministic(t *testing.T) {
	photo := image.NewRGBA(image.Rect(0, 0, 400, 300))
	a, err := Encode(photo, CardFor(lodgedFacts()))
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	b, err := Encode(photo, CardFor(lodgedFacts()))
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Error("two runs produced different bytes")
	}
}

func looksLikeCoordinate(s string) bool {
	for _, f := range strings.Fields(s) {
		f = strings.Trim(f, ",·()")
		if !strings.Contains(f, ".") {
			continue
		}
		parts := strings.SplitN(f, ".", 2)
		if len(parts) != 2 || len(parts[1]) < 2 {
			continue
		}
		allDigits := func(x string) bool {
			if x == "" {
				return false
			}
			for _, r := range strings.TrimPrefix(x, "-") {
				if r < '0' || r > '9' {
					return false
				}
			}
			return true
		}
		if allDigits(parts[0]) && allDigits(parts[1]) {
			return true
		}
	}
	return false
}
