package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// checkXformBitmap samples the mapped center of each transformed fill against
// the bitmap painter, away from the antialiased edges.
func checkXformBitmap(d *layout.Display, img *ebiten.Image) error {
	want, err := bitmap.Paint(d)
	if err != nil {
		return err
	}

	pixels := make([]byte, 240*140*4)
	img.ReadPixels(pixels)
	ppt := d.PixelPerPoint

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != layout.DisplayOpFillRect || !op.XformSet {
			continue
		}

		m := op.Transform()
		cx, cy := m.Apply(op.X+op.W/2, op.Y+op.H/2)
		x, y := int(cx*ppt), int(cy*ppt)

		if err := closePixel(pixels, want, 240, x, y); err != nil {
			return fmt.Errorf("transformed fill center: %w", err)
		}
	}

	return closePixel(pixels, want, 240, 232, 6)
}

// checkXformRepeat draws the page twice and requires bit-identical frames, so
// the cached raster blits stay deterministic across draws.
func checkXformRepeat(d *layout.Display) error {
	one, two := newParityImage(240, 140), newParityImage(240, 140)
	defer one.Deallocate()
	defer two.Deallocate()

	for _, img := range []*ebiten.Image{one, two} {
		for i := range d.Order {
			drawOp(img, &d.Ops[i], 0, 0)
		}
	}

	a, b := make([]byte, 240*140*4), make([]byte, 240*140*4)
	one.ReadPixels(a)
	two.ReadPixels(b)

	if !bytes.Equal(a, b) {
		return fmt.Errorf("repeated draw differs in %d channels: %s", countDiff(a, b), diffPixels(a, b, 240))
	}

	return nil
}

// checkXformParity draws a transformed page through drawOp and DrawVisible
// and requires identical pixels, then checks fills against the bitmap and
// the rotated run's placement.
func checkXformParity() error {
	source := `<body style="margin:0;height:140px;background:#f4f4f4">` +
		`<div style="position:absolute;left:20px;top:20px;width:100px;height:50px;background:#2860c8;transform:rotate(25deg)"></div>` +
		`<div style="position:absolute;left:150px;top:20px;width:60px;height:30px;background:#2ca04a;transform:skewX(20deg) scale(1.2,1)"></div>` +
		`<div style="position:absolute;left:20px;top:80px;width:90px;height:36px;border:3px solid #204020;transform:rotate(-15deg)"></div>` +
		`</body>`

	d, err := render.DisplayList(context.Background(), source, 240, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("transformed page should replay")
	}

	if err := checkXformRepeat(d); err != nil {
		return err
	}

	before, after := newParityImage(240, 140), newParityImage(240, 140)
	defer before.Deallocate()
	defer after.Deallocate()

	for i := range d.Order {
		drawOp(before, &d.Ops[i], 0, 0)
	}

	DrawVisible(after, d, 0, 0)

	a, b := make([]byte, 240*140*4), make([]byte, 240*140*4)
	before.ReadPixels(a)
	after.ReadPixels(b)

	if !bytes.Equal(a, b) {
		return fmt.Errorf("transformed draw differs in %d channels", countDiff(a, b))
	}

	if err := checkXformBitmap(d, after); err != nil {
		return err
	}

	return checkXformText()
}

// checkXformText pins a rotated run's placement: its ink must lie inside the
// bounds the matrix maps and nowhere far outside them.
func checkXformText() error {
	source := `<body style="margin:0;height:140px;background:#ffffff">` +
		`<div style="position:absolute;left:40px;top:40px;transform:rotate(20deg)">` +
		`<span style="color:#c00000;font-size:20px">rotated run</span></div></body>`

	d, err := render.DisplayList(context.Background(), source, 240, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("transformed text should replay")
	}

	img := newParityImage(240, 140)
	defer img.Deallocate()

	DrawVisible(img, d, 0, 0)

	pixels := make([]byte, 240*140*4)
	img.ReadPixels(pixels)

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != layout.DisplayOpText || !op.XformSet {
			continue
		}

		box, ok := PaintBounds(op, d.PointsPerPixel)
		if !ok || box.Empty() {
			return fmt.Errorf("transformed text is unbounded")
		}

		outside := box.Inset(-8)
		ink, stray := 0, 0

		for y := range 140 {
			for x := range 240 {
				at := (y*240 + x) * 4
				if pixels[at] == 255 && pixels[at+1] == 255 && pixels[at+2] == 255 {
					continue
				}

				if image.Pt(x, y).In(box) {
					ink++
				} else if !image.Pt(x, y).In(outside) {
					stray++
				}
			}
		}

		if ink == 0 {
			return fmt.Errorf("rotated run painted no ink inside %v", box)
		}
		if stray > 0 {
			return fmt.Errorf("rotated run painted %d pixels outside %v", stray, box)
		}

		return nil
	}

	return fmt.Errorf("no transformed text op")
}
