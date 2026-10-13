package redact

import (
	"context"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	pigo "github.com/esimov/pigo/core"
)

// Detector looks for one kind of thing in a photograph.
//
// It reports its own name so that the redaction record can say what ran. A
// detector that cannot name itself is treated as not having run at all, which
// is why Tool is part of the interface rather than a field somewhere.
type Detector interface {
	Kind() Kind
	Tool() string
	Detect(ctx context.Context, img image.Image) ([]Region, error)
}

// Scanner runs every detector over one photograph and records what each did.
//
// It returns a Scan rather than a list of regions because the interesting case
// is the detector that did not run: a scan that found nothing and a scan that
// never happened produce identical region lists, and only one of them is safe
// to publish.
type Scanner struct {
	detectors []Detector
}

func NewScanner(d ...Detector) *Scanner {
	return &Scanner{detectors: d}
}

// Scan never returns an error. A detector failing is a fact about the
// photograph that Decide has to weigh, not a reason to abandon the other
// detectors, and swallowing it is impossible because the failure is recorded
// against its kind.
func (s *Scanner) Scan(ctx context.Context, img image.Image) Scan {
	var out Scan
	for _, d := range s.detectors {
		c := Check{Kind: d.Kind(), Tool: d.Tool()}

		if err := ctx.Err(); err != nil {
			c.Err = fmt.Errorf("%s: %w", d.Tool(), err)
			out.Checks = append(out.Checks, c)
			continue
		}

		regions, err := d.Detect(ctx, img)
		if err != nil {
			c.Err = fmt.Errorf("%s: %w", d.Tool(), err)
		}
		out.Checks = append(out.Checks, c)
		out.Regions = append(out.Regions, regions...)
	}
	return out
}

// cascades holds the detector's weights. They are embedded rather than read
// from disk so that the binary cannot be deployed without them: a missing
// cascade would otherwise surface as a detector that quietly finds no faces.
//
// facefinder is from pigo (MIT, see cascade/LICENSE-pigo). ADR 0008 requires a
// permissive licence because the platform is distributed under AGPL.
//
//go:embed cascade/facefinder
var cascades embed.FS

var faceCascade = mustRead("cascade/facefinder")

func mustRead(name string) []byte {
	b, err := cascades.ReadFile(name)
	if err != nil {
		panic("redact: embedded cascade missing: " + err.Error())
	}
	return b
}

// Detector parameters. None of these is a legal constant, so hard rule 6 does
// not apply; all of them are provisional until the adversarial set ADR 0008
// requires has been built and scored.
const (
	// workingSide caps the long side the cascade runs over. A phone photograph
	// is several thousand pixels across and the cascade's cost is in the pixel
	// count, three times over because of faceAngles.
	workingSide = 1280

	// minFacePixels is the smallest face looked for, measured on the working
	// copy. Below this a face is not identifiable at the sizes a public page
	// renders, and the search cost grows sharply.
	minFacePixels = 20

	shiftFactor  = 0.1
	scaleFactor  = 1.1
	iouThreshold = 0.2

	// faceQuality is the cascade score below which a detection is dropped.
	// Erring low costs a mosaic square on a patch of road; erring high costs a
	// published face, so this sits below pigo's own suggested 5.0.
	faceQuality = 4.0
)

// faceAngles are the rotations the cascade is run at, as pigo's fraction of a
// full turn. Upright, then roughly eighteen degrees each way, because a face
// in a street photograph is rarely level and the upright cascade alone misses
// a tilted head.
var faceAngles = []float64{0, 0.05, 0.95}

const faceTool = "pigo-1.4.6/facefinder"

// Faces finds faces with a locally held cascade. Nothing leaves the process.
type Faces struct {
	pg *pigo.Pigo
}

func NewFaces() (*Faces, error) {
	pg, err := pigo.NewPigo().Unpack(faceCascade)
	if err != nil {
		return nil, fmt.Errorf("unpack face cascade: %w", err)
	}
	return &Faces{pg: pg}, nil
}

func (f *Faces) Kind() Kind   { return Face }
func (f *Faces) Tool() string { return faceTool }

func (f *Faces) Detect(ctx context.Context, img image.Image) ([]Region, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	work, factor := shrink(img)
	wb := work.Bounds()
	if shortSide(wb) < minFacePixels {
		return nil, nil
	}

	params := pigo.CascadeParams{
		MinSize:     minFacePixels,
		MaxSize:     shortSide(wb),
		ShiftFactor: shiftFactor,
		ScaleFactor: scaleFactor,
		ImageParams: pigo.ImageParams{
			Pixels: grayOf(work),
			Rows:   wb.Dy(),
			Cols:   wb.Dx(),
			Dim:    wb.Dx(),
		},
	}

	var found []pigo.Detection
	for _, angle := range faceAngles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		found = append(found, f.pg.RunCascade(params, angle)...)
	}
	found = f.pg.ClusterDetections(found, iouThreshold)

	origin := img.Bounds().Min
	var out []Region
	for _, d := range found {
		if d.Q < faceQuality {
			continue
		}
		out = append(out, Region{
			Kind:  Face,
			Rect:  restore(box(d.Row, d.Col, d.Scale), factor, origin),
			Score: float64(d.Q),
		})
	}
	return out, nil
}

// box turns pigo's centre-and-size detection into a rectangle. An odd scale
// rounds outwards: a pixel of road is cheaper than a pixel of jaw.
func box(row, col, scale int) image.Rectangle {
	half := (scale + 1) / 2
	return image.Rect(col-half, row-half, col+half, row+half)
}

// restore puts a box found in the working copy back into the coordinates of
// the photograph Apply will write into.
func restore(r image.Rectangle, factor int, origin image.Point) image.Rectangle {
	return image.Rect(
		origin.X+r.Min.X*factor, origin.Y+r.Min.Y*factor,
		origin.X+r.Max.X*factor, origin.Y+r.Max.Y*factor,
	)
}

// shrink returns an origin-anchored working copy no longer than workingSide on
// its long side, and the integer factor it was reduced by.
//
// The factor is an integer so that a box found in the copy maps back to the
// original exactly, with no rounding to reason about. Reduction is by averaging
// each block, which is the right filter for downsampling and loses no face
// that minFacePixels would have found.
func shrink(src image.Image) (*image.NRGBA, int) {
	b := src.Bounds()

	factor := 1
	if long := longSide(b); long > workingSide {
		factor = (long + workingSide - 1) / workingSide
	}

	if factor == 1 {
		out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(out, out.Bounds(), src, b.Min, draw.Src)
		return out, 1
	}

	w, h := max(b.Dx()/factor, 1), max(b.Dy()/factor, 1)
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := image.Rect(
				b.Min.X+x*factor, b.Min.Y+y*factor,
				b.Min.X+(x+1)*factor, b.Min.Y+(y+1)*factor,
			).Intersect(b)
			out.SetNRGBA(x, y, meanAt(src, cell))
		}
	}
	return out, factor
}

func meanAt(src image.Image, r image.Rectangle) color.NRGBA {
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

func longSide(r image.Rectangle) int {
	if r.Dx() > r.Dy() {
		return r.Dx()
	}
	return r.Dy()
}
