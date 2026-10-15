package replay

import (
	"bytes"
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// checkRasterCache requires identical pixels; four rasters bake, not five.
func checkRasterCache() error {
	direct := ebiten.NewImage(256, 200)
	cached := ebiten.NewImage(256, 200)
	defer direct.Dispose()
	defer cached.Dispose()

	before := rasterEntries()
	ops := rasterParityOps()
	for i := range ops {
		drawDirect(&ops[i], direct)
		drawOp(cached, &ops[i], 0, 0)
	}

	a, b := make([]byte, 256*200*4), make([]byte, 256*200*4)
	direct.ReadPixels(a)
	cached.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("raster cache differs in %d channels", countDiff(a, b))
	}

	if got := rasterEntries() - before; got != 4 {
		return fmt.Errorf("raster cache added %d entries, want 4", got)
	}

	return nil
}

// rasterEntries counts cached rasters.
func rasterEntries() int {
	rasters.mu.Lock()
	defer rasters.mu.Unlock()

	return len(rasters.entries)
}

// countDiff counts differing bytes.
func countDiff(a, b []byte) int {
	count := 0
	for i := range a {
		if a[i] != b[i] {
			count++
		}
	}

	return count
}

// checkRasterClip pins that a partial repaint clips like the direct draw.
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
