package redact

import (
	"context"
	"image"
	"sort"
)

// Plate detector parameters. Provisional until the adversarial set ADR 0008
// requires exists; every one of them trades a mosaic square on a patch of road
// against a readable number plate, so each errs towards the square.
const (
	plateTool = "edge-density-1"

	// A plate presents as a tight horizontal run of strong vertical edges. The
	// shape filters below are in working-copy pixels.
	minPlateWidth  = 24
	minPlateHeight = 8

	// Indian plates run from about 2:1 for the two-row format to about 4.7:1
	// for the single-row one. The window is widened at both ends because a
	// plate seen at an angle is shorter than a plate seen square on.
	minPlateAspect = 1.6
	maxPlateAspect = 7.0

	// edgeThreshold is the Sobel magnitude at which a pixel counts as an edge.
	edgeThreshold = 56

	// closeGap is how far apart two edge columns may be and still be read as
	// one run of characters.
	closeGap = 8

	// rowGap closes the gap between the rows of one plate, so the characters
	// and the plate's own border become a single region.
	rowGap = 2

	// minPlateFill is the share of a candidate box that must be edge after
	// closing. A sparse scatter of gravel fills far less than lettering does.
	minPlateFill = 0.45

	// maxPlateRegions keeps a busy street scene from coming back as a hundred
	// mosaic squares. The highest-scoring candidates are kept.
	maxPlateRegions = 24
)

// Plates finds number-plate-shaped regions with an edge-density localiser.
//
// It is deliberately not a reader: nothing here recognises characters, and the
// image never leaves the process. A localiser biased towards over-detection is
// the right shape for the job, because a false region costs a blurred patch of
// road and a missed one costs a published registration number.
//
// Its recall is unmeasured. Decide therefore holds the image whenever a second
// reading of the same photograph contradicts it, and the adversarial set ADR
// 0008 requires is still outstanding.
type Plates struct{}

func NewPlates() *Plates { return &Plates{} }

func (p *Plates) Kind() Kind   { return Plate }
func (p *Plates) Tool() string { return plateTool }

func (p *Plates) Detect(ctx context.Context, img image.Image) ([]Region, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	work, factor := shrink(img)
	b := work.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < minPlateWidth || h < minPlateHeight {
		return nil, nil
	}

	// Blur before differentiating. Sobel on raw pixels turns sensor noise and
	// road texture into edges everywhere, and the plate's own contrast is far
	// too strong to be lost to a 3x3 average.
	gray := blur3(grayOf(work), w, h)

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	mask := verticalEdges(gray, w, h)
	closeRuns(mask, w, h)

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	origin := img.Bounds().Min
	var out []Region
	for _, c := range components(mask, w, h) {
		if !plausiblePlate(c.box) {
			continue
		}
		fill := float64(c.size) / float64(c.box.Dx()*c.box.Dy())
		if fill < minPlateFill {
			continue
		}
		out = append(out, Region{
			Kind:  Plate,
			Rect:  restore(c.box, factor, origin),
			Score: fill,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > maxPlateRegions {
		out = out[:maxPlateRegions]
	}
	return out, nil
}

// plausiblePlate rejects shapes no number plate has. It is the cheapest filter
// available and it runs before anything measures density.
func plausiblePlate(r image.Rectangle) bool {
	w, h := r.Dx(), r.Dy()
	if w < minPlateWidth || h < minPlateHeight {
		return false
	}
	aspect := float64(w) / float64(h)
	return aspect >= minPlateAspect && aspect <= maxPlateAspect
}

func grayOf(img *image.NRGBA) []uint8 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	g := make([]uint8, w*h)
	for y := 0; y < h; y++ {
		row := img.Pix[y*img.Stride:]
		for x := 0; x < w; x++ {
			g[y*w+x] = uint8((299*uint32(row[x*4]) +
				587*uint32(row[x*4+1]) +
				114*uint32(row[x*4+2])) / 1000)
		}
	}
	return g
}

func blur3(src []uint8, w, h int) []uint8 {
	out := make([]uint8, len(src))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sum, n uint32
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := x+dx, y+dy
					if nx < 0 || ny < 0 || nx >= w || ny >= h {
						continue
					}
					sum += uint32(src[ny*w+nx])
					n++
				}
			}
			out[y*w+x] = uint8(sum / n)
		}
	}
	return out
}

// verticalEdges marks pixels where the horizontal gradient is strong. A plate's
// characters are vertical strokes, so the x gradient is what separates them
// from a kerb, a wire or a horizon.
func verticalEdges(gray []uint8, w, h int) []bool {
	mask := make([]bool, w*h)
	at := func(x, y int) int32 {
		if x < 0 {
			x = 0
		}
		if x >= w {
			x = w - 1
		}
		if y < 0 {
			y = 0
		}
		if y >= h {
			y = h - 1
		}
		return int32(gray[y*w+x])
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			g := -at(x-1, y-1) + at(x+1, y-1) +
				-2*at(x-1, y) + 2*at(x+1, y) +
				-at(x-1, y+1) + at(x+1, y+1)
			if g < 0 {
				g = -g
			}
			mask[y*w+x] = g/4 >= edgeThreshold
		}
	}
	return mask
}

// closeRuns joins edges that are close enough together to be one piece of
// lettering: first along each row, then between neighbouring rows, so a plate's
// characters and border become one region instead of twenty.
func closeRuns(mask []bool, w, h int) {
	for y := 0; y < h; y++ {
		last := -1
		for x := 0; x < w; x++ {
			if !mask[y*w+x] {
				continue
			}
			if last >= 0 && x-last <= closeGap+1 {
				for f := last + 1; f < x; f++ {
					mask[y*w+f] = true
				}
			}
			last = x
		}
	}

	for x := 0; x < w; x++ {
		last := -1
		for y := 0; y < h; y++ {
			if !mask[y*w+x] {
				continue
			}
			if last >= 0 && y-last <= rowGap+1 {
				for f := last + 1; f < y; f++ {
					mask[f*w+x] = true
				}
			}
			last = y
		}
	}
}

type component struct {
	box  image.Rectangle
	size int
}

// components labels the mask by flood fill. Four-connectivity is enough once
// closeRuns has joined what belongs together, and it keeps diagonal road
// texture from chaining into one region covering the frame.
func components(mask []bool, w, h int) []component {
	seen := make([]bool, len(mask))
	var out []component
	stack := make([]int, 0, 1024)

	for i := range mask {
		if !mask[i] || seen[i] {
			continue
		}
		seen[i] = true
		stack = append(stack[:0], i)

		minX, minY := w, h
		maxX, maxY, size := -1, -1, 0

		for len(stack) > 0 {
			p := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			x, y := p%w, p/w
			size++

			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}

			if x > 0 && mask[p-1] && !seen[p-1] {
				seen[p-1] = true
				stack = append(stack, p-1)
			}
			if x < w-1 && mask[p+1] && !seen[p+1] {
				seen[p+1] = true
				stack = append(stack, p+1)
			}
			if y > 0 && mask[p-w] && !seen[p-w] {
				seen[p-w] = true
				stack = append(stack, p-w)
			}
			if y < h-1 && mask[p+w] && !seen[p+w] {
				seen[p+w] = true
				stack = append(stack, p+w)
			}
		}

		out = append(out, component{
			box:  image.Rect(minX, minY, maxX+1, maxY+1),
			size: size,
		})
	}
	return out
}
