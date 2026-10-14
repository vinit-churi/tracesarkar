package share

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"

	xdraw "golang.org/x/image/draw"
)

// Shape is the frame a card is drawn into.
type Shape string

const (
	// Feed is the default: the photograph at its own width with the facts
	// added below it.
	Feed Shape = "feed"
	// Story is 9:16, for WhatsApp Status and Instagram Stories — a real MMR
	// channel, and one that letterboxes a landscape card into illegibility.
	Story Shape = "story"
)

// Shapes lists what can be drawn.
func Shapes() []Shape { return []Shape{Feed, Story} }

// KnownShape reports whether a layout exists for this shape. An unknown shape
// is refused rather than falling back to the default, because a caller asking
// for a frame it will post into wants that frame or an error.
func KnownShape(s Shape) bool {
	for _, k := range Shapes() {
		if k == s {
			return true
		}
	}
	return false
}

// storyWidth is the canvas width. 1080 is what both Instagram and WhatsApp
// Status serve at, so drawing wider only costs bytes on a 2G connection.
const storyWidth = 1080

// DrawStory fits the card into a 9:16 frame.
//
// The photograph keeps its aspect ratio and is centred in the space left above
// the fact block. Stretching it to fill the frame would be the one edit that
// makes the evidence untrue, and cropping it could remove the defect being
// reported — so the frame is padded instead, in the same paper the block sits
// on.
func DrawStory(photo image.Image, c Card) (image.Image, error) {
	if photo == nil {
		return nil, fmt.Errorf("share: no photograph to annotate")
	}
	pb := photo.Bounds()
	if pb.Dx() <= 0 || pb.Dy() <= 0 {
		return nil, fmt.Errorf("share: photograph has no area")
	}

	// The block is drawn first, at the canvas width, so its real height is
	// known before the photograph is given the space that is left.
	block, err := Draw(image.NewRGBA(image.Rect(0, 0, storyWidth, 1)), c)
	if err != nil {
		return nil, err
	}
	blockH := block.Bounds().Dy() - 1

	height := storyWidth * 16 / 9
	room := height - blockH
	if room < storyWidth/2 {
		// A block taller than the frame leaves nothing for the photograph. Let
		// the canvas grow rather than crop the facts: the ratio is a target,
		// and a card missing a fact is worse than one slightly too tall.
		room = storyWidth / 2
		height = room + blockH
	}

	scale := float64(room) / float64(pb.Dy())
	if w := float64(storyWidth) / float64(pb.Dx()); w < scale {
		scale = w
	}
	fitW, fitH := int(float64(pb.Dx())*scale), int(float64(pb.Dy())*scale)

	canvas := image.NewRGBA(image.Rect(0, 0, storyWidth, height))
	xdraw.Draw(canvas, canvas.Bounds(), &image.Uniform{cardPaper}, image.Point{}, xdraw.Src)

	at := image.Rect((storyWidth-fitW)/2, (room-fitH)/2, (storyWidth-fitW)/2+fitW, (room-fitH)/2+fitH)
	xdraw.CatmullRom.Scale(canvas, at, photo, pb, xdraw.Over, nil)

	// The block is redrawn over a one-pixel strip so it lands at the bottom
	// with exactly the layout Draw produces, rather than being reimplemented.
	xdraw.Draw(canvas, image.Rect(0, room, storyWidth, height),
		block, image.Point{X: 0, Y: 1}, xdraw.Src)
	return canvas, nil
}

// EncodeShape draws a card in the requested frame and fits it to the budget.
func EncodeShape(photo image.Image, c Card, s Shape) ([]byte, error) {
	var img image.Image
	var err error
	switch s {
	case Feed:
		img, err = Draw(photo, c)
	case Story:
		img, err = DrawStory(photo, c)
	default:
		return nil, fmt.Errorf("share: no layout for shape %q", s)
	}
	if err != nil {
		return nil, err
	}
	return fitToBudget(img)
}

func fitToBudget(img image.Image) ([]byte, error) {
	var last []byte
	for _, q := range []int{82, 70, 60, 50, 40} {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
			return nil, fmt.Errorf("share: encode card: %w", err)
		}
		last = buf.Bytes()
		if len(last) <= CardBudgetBytes {
			return last, nil
		}
	}
	return last, nil
}
