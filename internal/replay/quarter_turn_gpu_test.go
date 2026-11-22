package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// quarterTurnDegrees are the glyph rotations vertical writing can emit.
var quarterTurnDegrees = []float32{90, -90, 180, 270}

// checkQuarterTurnText pins quarter-turn text replay: a rotated run draws
// exactly like the same run with the baked rotation, a clipped partial redraw
// matches the full draw inside the clip, and a real vertical-rl page puts its
// ink where the software painter does.
func checkQuarterTurnText() error {
	ops, err := textParityOps()
	if err != nil {
		return err
	}
	for _, deg := range quarterTurnDegrees {
		if err := checkQuarterTurnBaked(ops[0], deg); err != nil {
			return err
		}
		if err := checkQuarterTurnPartial(ops[0], deg); err != nil {
			return err
		}
	}

	return checkVerticalPageInk()
}

// quarterTurnRun places the parity run at the canvas center with deg.
func quarterTurnRun(base layout.DisplayOp, deg float32) layout.DisplayOp {
	op := base
	op.X, op.Y = 160, 120
	op.RotateDeg = deg
	op.XformSet = false

	return op
}

func checkQuarterTurnBaked(base layout.DisplayOp, deg float32) error {
	rotated := quarterTurnRun(base, deg)
	m, ok := masking.TextMatrix(&rotated)
	if !ok {
		return fmt.Errorf("rotate %v: no text matrix", deg)
	}
	baked := rotated
	baked.RotateDeg = 0
	baked.SetXform(m)
	baked.XformSet = true

	var renderer Renderer
	defer renderer.Close()
	actual, expected := newParityImage(320, 240), newParityImage(320, 240)
	defer actual.Deallocate()
	defer expected.Deallocate()
	renderer.drawOp(actual, &rotated, 0, 0)
	renderer.drawOp(expected, &baked, 0, 0)
	a, b := make([]byte, 320*240*4), make([]byte, 320*240*4)
	actual.ReadPixels(a)
	expected.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("rotate %v: rotated run differs from baked rotation in %d channels", deg, countDiff(a, b))
	}
	if bytes.Equal(a, make([]byte, len(a))) {
		return fmt.Errorf("rotate %v: rotated run drew nothing", deg)
	}
	got, gotOK := opBounds(&rotated, 1/pxPerPt)
	want, wantOK := opBounds(&baked, 1/pxPerPt)
	if got != want || gotOK != wantOK {
		return fmt.Errorf("rotate %v: bounds = %v,%t, want %v,%t", deg, got, gotOK, want, wantOK)
	}

	return nil
}

func checkQuarterTurnPartial(base layout.DisplayOp, deg float32) error {
	rotated := quarterTurnRun(base, deg)
	upright := base
	upright.X, upright.Y = 10, 30
	display := &layout.Display{
		Width: 320, Height: 240, PointsPerPixel: 1 / pxPerPt,
		Ops: []layout.DisplayOp{upright, rotated}, Order: []int{0, 1},
	}
	box, ok := opBounds(&rotated, display.PointsPerPixel)
	if !ok || box.Empty() {
		return fmt.Errorf("rotate %v: unbounded run", deg)
	}
	// Cut the run's box in half along its long side so the clip slices ink.
	rect := box
	if box.Dy() >= box.Dx() {
		rect.Max.Y = (box.Min.Y + box.Max.Y) / 2
	} else {
		rect.Max.X = (box.Min.X + box.Max.X) / 2
	}

	full, partial := newParityImage(320, 240), newParityImage(320, 240)
	defer full.Deallocate()
	defer partial.Deallocate()
	DrawVisible(full, display, 0, 0)
	DrawRect(partial, display, rect, 0, 0)
	a, b := make([]byte, 320*240*4), make([]byte, 320*240*4)
	full.ReadPixels(a)
	partial.ReadPixels(b)
	diff, inked := 0, false
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			at := (y*320 + x) * 4
			if a[at+3] != 0 {
				inked = true
			}
			if !bytes.Equal(a[at:at+4], b[at:at+4]) {
				diff++
			}
		}
	}
	if !inked {
		return fmt.Errorf("rotate %v: clip %v holds no ink", deg, rect)
	}
	if diff != 0 {
		return fmt.Errorf("rotate %v: partial redraw in %v differs from the full draw at %d pixels", deg, rect, diff)
	}

	return nil
}

// checkVerticalPageInk compares the ink box of a vertical-rl page between
// replay and the software painter.
func checkVerticalPageInk() error {
	source := `<body style="margin:0;background:#fff"><div style="position:absolute;left:60px;top:20px;` +
		`width:60px;height:150px;color:#000;font-size:18px;writing-mode:vertical-rl">Vertical text</div></body>`
	d, err := render.DisplayList(context.Background(), source, 240, 200)
	if err != nil {
		return err
	}
	if !render.Replayable(d) {
		return fmt.Errorf("vertical-rl page should replay")
	}
	soft, err := bitmap.Paint(d)
	if err != nil {
		return err
	}
	img := newParityImage(240, 200)
	defer img.Deallocate()
	DrawVisible(img, d, 0, 0)
	pixels := make([]byte, 240*200*4)
	img.ReadPixels(pixels)

	var gpuInk, softInk image.Rectangle
	for y := 0; y < 200; y++ {
		for x := 0; x < 240; x++ {
			at := (y*240 + x) * 4
			if pixels[at+3] >= 128 && (pixels[at] < 200 || pixels[at+1] < 200 || pixels[at+2] < 200) {
				gpuInk = gpuInk.Union(image.Rect(x, y, x+1, y+1))
			}
			r, g, b, alpha := soft.At(x, y).RGBA()
			if alpha >= 0x8000 && (r < 0xc800 || g < 0xc800 || b < 0xc800) {
				softInk = softInk.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	if gpuInk.Empty() || softInk.Empty() {
		return fmt.Errorf("vertical-rl ink missing: replay %v, software %v", gpuInk, softInk)
	}
	for _, e := range [4][2]int{
		{gpuInk.Min.X, softInk.Min.X}, {gpuInk.Min.Y, softInk.Min.Y},
		{gpuInk.Max.X, softInk.Max.X}, {gpuInk.Max.Y, softInk.Max.Y},
	} {
		if d := e[0] - e[1]; d < -3 || d > 3 {
			return fmt.Errorf("vertical-rl ink box: replay %v, software %v", gpuInk, softInk)
		}
	}

	return nil
}
