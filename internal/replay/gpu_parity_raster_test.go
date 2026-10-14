package replay

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// checkRasterCache draws repeated rounded shapes through the direct path and
// through the raster cache and requires identical pixels. Two of the five ops
// share one shape at different integer positions, so the check must bake four
// rasters, not five.
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

// rasterEntries reports the number of cached rasters.
func rasterEntries() int {
	rasters.mu.Lock()
	defer rasters.mu.Unlock()

	return len(rasters.entries)
}

// countDiff counts differing bytes between two pixel buffers.
func countDiff(a, b []byte) int {
	count := 0
	for i := range a {
		if a[i] != b[i] {
			count++
		}
	}

	return count
}
