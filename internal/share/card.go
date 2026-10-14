package share

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// CardBudgetBytes is what one card may weigh.
//
// The public page is costed at 150 KB total on a 2G connection and the card is
// the heaviest thing on it, so it gets most of that and nothing more. A card
// that will not load in a Mumbai suburb at 7 p.m. is not a card.
const CardBudgetBytes = 140 << 10

// Contract is a published contract attribution. It is a pointer on Facts
// because today there is nothing to put in it: the attribution exists in
// internal/attribute but is not yet published on a public surface, and a
// struct that is always present invites a blank line on the card.
//
// Nothing here is printed without SourceRef. Hard rule 2 has no exception for
// a field that happens to be empty.
type Contract struct {
	Code         string
	Contractor   string
	WarrantyTill *time.Time
	SourceRef    string
}

// Card is the fact block burned into the image.
//
// It is text, decided here and rendered separately, so that what the card
// claims can be tested without rendering a pixel — and so the same block can
// be read out to a screen reader on the page.
type Card struct {
	Headline string
	Rows     [][2]string
	Footer   string
	Mark     string
}

// CardFor derives the block from the fact set.
func CardFor(f Facts) Card {
	en, err := vocabulary(EN)
	thing := strings.TrimSpace(f.Subcategory)
	if err == nil {
		if t, _ := en.thing(f); t != "" {
			thing = t
		}
	}

	c := Card{
		Headline: strings.ToUpper(thing),
		Mark:     "TraceSarkar",
		Footer:   strings.TrimPrefix(strings.TrimPrefix(f.Permalink, "https://"), "http://"),
	}
	if f.Ward != "" {
		c.Headline += "  ·  " + f.Ward + " ward"
	}
	if f.Authority != "" {
		c.Rows = append(c.Rows, [2]string{"Authority", f.Authority + " · " + f.Ward + " ward"})
	}
	// No coordinate, coarsened or otherwise. The page shows the cell on a map
	// with the wording that explains it; a bare pair of numbers on a forwarded
	// image loses that wording and keeps the precision.
	if !f.FirstSeen.IsZero() {
		c.Rows = append(c.Rows, [2]string{"First seen", f.FirstSeen.UTC().Format("2 Jan 2006")})
	}
	if f.Sightings > 1 {
		c.Rows = append(c.Rows, [2]string{"Seen", strconv.Itoa(f.Sightings) + " times"})
	}
	if f.Contract != nil && strings.TrimSpace(f.Contract.SourceRef) != "" {
		if f.Contract.Code != "" {
			c.Rows = append(c.Rows, [2]string{"Contract", f.Contract.Code})
		}
		if f.Contract.WarrantyTill != nil {
			c.Rows = append(c.Rows, [2]string{"Warranty",
				"valid until " + f.Contract.WarrantyTill.UTC().Format("2 Jan 2006")})
		}
	}
	if f.Lodged && f.LodgedOn != nil {
		c.Rows = append(c.Rows, [2]string{"Authority told", f.LodgedOn.UTC().Format("2 Jan 2006")})
		if f.DeadlineHours > 0 && strings.TrimSpace(f.DeadlineShortRef) != "" {
			v := strconv.Itoa(f.DeadlineHours) + " hours · " + f.DeadlineShortRef
			if f.DeadlineState == "elapsed" {
				v += " · that time has passed"
			}
			c.Rows = append(c.Rows, [2]string{"To attend", v})
		}
	} else {
		c.Rows = append(c.Rows, [2]string{"Authority told", "no record of it"})
	}
	if l := cardStatus(f.Status); l != "" {
		c.Rows = append(c.Rows, [2]string{"Status", l})
	}
	return c
}

// cardStatus is the card's own wording, short enough for a fact block.
//
// claimed_resolved and citizen_confirmed are different claims by different
// parties (G14), so they get different words here as everywhere else, and an
// unknown status is printed as unknown rather than guessed at.
func cardStatus(s string) string {
	switch s {
	case "open":
		return "open"
	case "claimed_resolved":
		return "authority recorded it attended; no resident photograph of a fix"
	case "citizen_confirmed":
		return "a resident photographed it fixed"
	case "reopened":
		return "seen again after being recorded attended"
	case "":
		return ""
	}
	return "not published"
}

// Lines is the whole block as text, in reading order. It is what the tone
// check runs over and what a screen reader would be given.
func (c Card) Lines() []string {
	out := []string{c.Headline}
	for _, r := range c.Rows {
		out = append(out, r[0]+"  "+r[1])
	}
	return append(out, c.Footer+"  ·  "+c.Mark+"  ·  sources on the page")
}

// Layout constants. The block is sized from the photograph's width so a card
// looks the same whatever the camera produced.
const (
	cardPad      = 28
	cardRowGap   = 34
	cardLabelCol = 150
)

var (
	cardInk   = color.RGBA{R: 0x12, G: 0x14, B: 0x18, A: 0xff}
	cardMuted = color.RGBA{R: 0x5a, G: 0x60, B: 0x6a, A: 0xff}
	cardPaper = color.RGBA{R: 0xf7, G: 0xf6, B: 0xf3, A: 0xff}
	cardRule  = color.RGBA{R: 0xd8, G: 0xd5, B: 0xcf, A: 0xff}
)

// Draw puts the fact block under the photograph.
//
// Under, not over: the photograph is the evidence, and a caption box across it
// covers the thing being reported. It also means no text sits on top of a
// region a detector mosaicked, which would make the mosaic look like a
// deliberate redaction of the facts.
func Draw(photo image.Image, c Card) (image.Image, error) {
	if photo == nil {
		return nil, fmt.Errorf("share: no photograph to annotate")
	}
	pb := photo.Bounds()
	if pb.Dx() <= 0 || pb.Dy() <= 0 {
		return nil, fmt.Errorf("share: photograph has no area")
	}

	scale := float64(pb.Dx()) / 900.0
	if scale < 0.6 {
		scale = 0.6
	}
	if scale > 2.0 {
		scale = 2.0
	}
	head, err := face(gobold.TTF, 20*scale)
	if err != nil {
		return nil, err
	}
	defer head.Close()
	label, err := face(goregular.TTF, 13*scale)
	if err != nil {
		return nil, err
	}
	defer label.Close()
	value, err := face(gobold.TTF, 15*scale)
	if err != nil {
		return nil, err
	}
	defer value.Close()
	foot, err := face(goregular.TTF, 12*scale)
	if err != nil {
		return nil, err
	}
	defer foot.Close()

	pad := int(cardPad * scale)
	gap := int(cardRowGap * scale)
	blockH := pad + int(30*scale) + pad/2 + len(c.Rows)*gap + pad/2 + int(22*scale) + pad
	out := image.NewRGBA(image.Rect(0, 0, pb.Dx(), pb.Dy()+blockH))

	draw.Draw(out, image.Rect(0, 0, pb.Dx(), pb.Dy()), photo, pb.Min, draw.Src)
	draw.Draw(out, image.Rect(0, pb.Dy(), pb.Dx(), out.Bounds().Dy()),
		&image.Uniform{cardPaper}, image.Point{}, draw.Src)

	y := pb.Dy() + pad + int(20*scale)
	text(out, head, cardInk, pad, y, c.Headline)
	y += pad / 2

	// A rule under the headline, so a forwarded screenshot cropped at the
	// wrong place still reads as two zones rather than one paragraph.
	draw.Draw(out, image.Rect(pad, y, pb.Dx()-pad, y+max(1, int(scale))),
		&image.Uniform{cardRule}, image.Point{}, draw.Src)

	y += gap
	for _, r := range c.Rows {
		text(out, label, cardMuted, pad, y, r[0])
		text(out, value, cardInk, pad+int(cardLabelCol*scale), y, r[1])
		y += gap
	}

	y += pad / 2
	text(out, foot, cardMuted, pad, y, c.Footer+"   ·   sources on the page")
	mark := c.Mark
	w := font.MeasureString(foot, mark).Round()
	text(out, foot, cardInk, pb.Dx()-pad-w, y, mark)
	return out, nil
}

// Encode draws the card and compresses it into the transfer budget.
//
// Quality is stepped down rather than the image being cropped or the block
// shrunk: a card that has lost a fact is worse than a card that is slightly
// soft, and the facts are the point.
func Encode(photo image.Image, c Card) ([]byte, error) {
	annotated, err := Draw(photo, c)
	if err != nil {
		return nil, err
	}
	var last []byte
	for _, q := range []int{82, 70, 60, 50, 40} {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, annotated, &jpeg.Options{Quality: q}); err != nil {
			return nil, fmt.Errorf("share: encode card: %w", err)
		}
		last = buf.Bytes()
		if len(last) <= CardBudgetBytes {
			return last, nil
		}
	}
	return last, nil
}

func face(ttf []byte, size float64) (interface {
	font.Face
	Close() error
}, error) {
	f, err := opentype.Parse(ttf)
	if err != nil {
		return nil, fmt.Errorf("share: parse card font: %w", err)
	}
	fc, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size: size, DPI: 72, Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("share: build card font: %w", err)
	}
	return fc, nil
}

func text(dst draw.Image, f font.Face, c color.Color, x, y int, s string) {
	d := &font.Drawer{
		Dst: dst, Src: &image.Uniform{c}, Face: f,
		Dot: fixed.P(x, y),
	}
	d.DrawString(s)
}
