package replay

import (
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
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
