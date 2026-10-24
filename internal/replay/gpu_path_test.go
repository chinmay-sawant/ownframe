package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"math"
	"strconv"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// The corner-shape pages below are 200 by 140 CSS pixels. The box is a 128 by
// 88 CSS pixel border box at (20, 20), a 24 px bevel or scoop corner, and a
// 4 px border. The body carries the height so the page fill reaches the
// bottom edge and every empty sample reads as the page colour.
const (
	pathPageWidth  = 200
	pathPageHeight = 140
	pathChannels   = 4
	pathSlack      = 3
)

// pathBoxStyle is the shared shaped box: a background, a uniform border, a
// 24 px radius, and one corner shape, over the given page colour.
func pathBoxStyle(backdrop, shape, extra string) string {
	return `<body style="margin:0;height:` + strconv.Itoa(pathPageHeight) + `px;background:` + backdrop + `">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:24px;corner-shape:` +
		shape + `;` + extra + `"></div></body>`
}

// checkPathGolden is the shaped-box golden: a bevel box must paint its fill,
// its border, the cut corners, and the page around it at exact colours. The
// samples are picked away from the antialiased edges.
func checkPathGolden() error {
	display, err := render.DisplayList(context.Background(), pathBoxStyle("#ffffff", "bevel", ""), pathPageWidth, pathPageHeight)
	if err != nil {
		return err
	}

	if !render.Replayable(display) {
		return fmt.Errorf("golden corner-shape page should replay")
	}

	img := newParityImage(pathPageWidth, pathPageHeight)
	defer img.Deallocate()

	DrawVisible(img, display, 0, 0)

	pixels := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	img.ReadPixels(pixels)

	samples := []struct {
		name string
		x, y int
		want [4]uint8
	}{
		{"fill centre", 84, 64, [4]uint8{48, 96, 192, 255}},
		{"fill under the top border", 84, 26, [4]uint8{48, 96, 192, 255}},
		{"top border", 84, 20, [4]uint8{16, 32, 48, 255}},
		{"left border", 20, 84, [4]uint8{16, 32, 48, 255}},
		{"bevel stroke", 32, 32, [4]uint8{16, 32, 48, 255}},
		{"inside the bevel", 36, 36, [4]uint8{48, 96, 192, 255}},
		{"outside the bevel", 28, 28, [4]uint8{255, 255, 255, 255}},
		{"right of the box", 160, 64, [4]uint8{255, 255, 255, 255}},
		{"below the box", 84, 120, [4]uint8{255, 255, 255, 255}},
		{"page corner", 4, 4, [4]uint8{255, 255, 255, 255}},
	}

	for _, sample := range samples {
		if err := checkPathPixel(pixels, sample.x, sample.y, sample.want); err != nil {
			return fmt.Errorf("%s: %w", sample.name, err)
		}
	}

	return nil
}

// checkPathXform draws a rotated scoop box twice, requires identical pixels,
// then requires the ink to leave the upright box (the matrix applied) and to
// stay inside the painted bounds (the matrix bounded).
func checkPathXform() error {
	display, err := render.DisplayList(context.Background(), pathBoxStyle("#ffffff", "scoop", "transform:rotate(15deg)"), pathPageWidth, pathPageHeight)
	if err != nil {
		return err
	}

	if !render.Replayable(display) {
		return fmt.Errorf("transformed corner-shape page should replay")
	}

	before, after := newParityImage(pathPageWidth, pathPageHeight), newParityImage(pathPageWidth, pathPageHeight)
	defer before.Deallocate()
	defer after.Deallocate()

	for _, index := range display.Order {
		drawOp(before, &display.Ops[index], 0, 0)
	}

	DrawVisible(after, display, 0, 0)

	one := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	two := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	before.ReadPixels(one)
	after.ReadPixels(two)

	if !bytes.Equal(one, two) {
		return fmt.Errorf("transformed path differs in %d channels: %s",
			countDiff(one, two), diffPixels(one, two, pathPageWidth))
	}

	op := findPathOp(display, false)
	if op == nil {
		return fmt.Errorf("transformed page has no fill path")
	}

	ppt := display.PointsPerPixel
	upright := image.Rect(
		int(math.Round(op.X/ppt)), int(math.Round(op.Y/ppt)),
		int(math.Round((op.X+op.W)/ppt)), int(math.Round((op.Y+op.H)/ppt)),
	)

	limit, ok := PaintBounds(op, ppt)
	if !ok || limit.Empty() {
		return fmt.Errorf("transformed path is unbounded")
	}

	limit = limit.Inset(-pathSlack)

	ink, outside, stray := 0, 0, 0

	for y := range pathPageHeight {
		for x := range pathPageWidth {
			at := (y*pathPageWidth + x) * pathChannels
			if two[at] == 255 && two[at+1] == 255 && two[at+2] == 255 {
				continue
			}

			ink++

			point := image.Pt(x, y)
			if !point.In(upright) {
				outside++
			}

			if !point.In(limit) {
				stray++
			}
		}
	}

	if ink == 0 {
		return fmt.Errorf("transformed path painted no ink")
	}

	if outside == 0 {
		return fmt.Errorf("a 15 degree rotation left no ink outside the upright box; the matrix was not applied")
	}

	if stray > 0 {
		return fmt.Errorf("transformed path painted %d pixels outside %v", stray, limit)
	}

	return nil
}

// checkPathGroup puts the shaped box inside a multiply group and requires the
// full and partial draws to agree inside the repainted rect, the composite to
// show the multiplied fill, and the cut corner to leave the backdrop alone.
func checkPathGroup() error {
	display, err := render.DisplayList(context.Background(), pathBoxStyle("#d2d2d2", "bevel", "mix-blend-mode:multiply"), pathPageWidth, pathPageHeight)
	if err != nil {
		return err
	}

	if !render.Replayable(display) {
		return fmt.Errorf("corner-shape group should replay")
	}

	full, part := newParityImage(pathPageWidth, pathPageHeight), newParityImage(pathPageWidth, pathPageHeight)
	defer full.Deallocate()
	defer part.Deallocate()

	DrawVisible(full, display, 0, 0)

	rect := image.Rect(40, 40, 120, 100)
	DrawRect(part, display, rect, 0, 0)

	one := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	two := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	full.ReadPixels(one)
	part.ReadPixels(two)

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		from, to := (y*pathPageWidth+rect.Min.X)*pathChannels, (y*pathPageWidth+rect.Max.X)*pathChannels
		if !bytes.Equal(one[from:to], two[from:to]) {
			return fmt.Errorf("shaped group partial repaint differs on row %d in %d channels",
				y, countDiff(one[from:to], two[from:to]))
		}
	}

	// multiply: #3060c0 over #d2d2d2, channel by channel.
	if err := checkPathPixel(one, 84, 64, [4]uint8{40, 79, 158, 255}); err != nil {
		return fmt.Errorf("multiplied fill: %w", err)
	}

	// The bevel cut is outside the shaped layer, so the group leaves the
	// backdrop untouched there.
	if err := checkPathPixel(one, 28, 28, [4]uint8{210, 210, 210, 255}); err != nil {
		return fmt.Errorf("backdrop under the cut corner: %w", err)
	}

	return checkPathPixel(one, 4, 4, [4]uint8{210, 210, 210, 255})
}

// checkPathPixel compares one pixel against a wanted colour within the test
// slack, which covers one step of shader rounding.
func checkPathPixel(pixels []byte, x, y int, want [4]uint8) error {
	at := (y*pathPageWidth + x) * pathChannels

	for index, channel := range want {
		if delta := int(pixels[at+index]) - int(channel); delta < -pathSlack || delta > pathSlack {
			return fmt.Errorf("pixel (%d,%d): got %v, want %v", x, y, pixels[at:at+pathChannels], want)
		}
	}

	return nil
}

// findPathOp returns the first corner path op of a display, by fill or stroke,
// or nil when the display carries none.
func findPathOp(display *layout.Display, stroke bool) *layout.DisplayOp {
	for index := range display.Ops {
		op := &display.Ops[index]
		if op.Kind != layout.DisplayOpPath {
			continue
		}

		if corner, ok := op.CornerPathValue(); ok && corner != nil && corner.Stroke == stroke {
			return op
		}
	}

	return nil
}
