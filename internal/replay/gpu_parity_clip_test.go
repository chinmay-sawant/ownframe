package replay

import (
	"bytes"
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// checkRasterClip draws a display's rounded shapes through DrawRect, which
// clips into a sub-image, and compares against the same shapes drawn
// directly into the same sub-image. The raster cache bakes unclipped rasters,
// so this pins that a partial repaint still clips the same way.
func checkRasterClip() error {
	ops := rasterParityOps()
	display := &layout.Display{
		Width: 256, Height: 200, PointsPerPixel: 72.0 / 96.0,
		Ops: ops, Order: []int{0, 1, 2, 3, 4},
	}
	rect := image.Rect(12, 16, 160, 120)

	direct := ebiten.NewImage(256, 200)
	cached := ebiten.NewImage(256, 200)
	defer direct.Dispose()
	defer cached.Dispose()

	clip, ok := direct.SubImage(rect).(*ebiten.Image)
	if !ok {
		return fmt.Errorf("sub-image")
	}
	for i := range ops {
		drawDirect(&ops[i], clip)
	}
	DrawRect(cached, display, rect, 0, 0)

	a, b := make([]byte, 256*200*4), make([]byte, 256*200*4)
	direct.ReadPixels(a)
	cached.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("clipped raster differs in %d channels", countDiff(a, b))
	}

	return nil
}
