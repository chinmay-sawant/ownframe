package replay

import (
	"context"
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

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

	img := ebiten.NewImage(240, 140)
	defer img.Dispose()

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
