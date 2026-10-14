package replay

import (
	"bytes"
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// checkTextClip draws the text parity runs through DrawRect, which clips into
// a sub-image, and compares against the same runs drawn directly into that
// sub-image. A baked run is not clipped at bake time, so this pins that a
// partial repaint clips it exactly as the direct glyphs would be, including
// the run past the size cap that stays on the direct path.
func checkTextClip() error {
	ops, err := textParityOps()
	if err != nil {
		return err
	}

	display := &layout.Display{
		Width: 320, Height: 240, PointsPerPixel: pxPerPt,
		Ops: ops, Order: []int{0, 1, 2, 3, 4, 5},
	}
	rect := image.Rect(8, 40, 220, 232)

	direct := ebiten.NewImage(320, 240)
	cached := ebiten.NewImage(320, 240)
	defer direct.Dispose()
	defer cached.Dispose()

	clip, ok := direct.SubImage(rect).(*ebiten.Image)
	if !ok {
		return fmt.Errorf("sub-image")
	}
	for i := range ops {
		drawTextDirect(&ops[i], clip)
	}
	DrawRect(cached, display, rect, 0, 0)

	a, b := make([]byte, 320*240*4), make([]byte, 320*240*4)
	direct.ReadPixels(a)
	cached.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("clipped text raster differs in %d channels", countDiff(a, b))
	}

	return nil
}
