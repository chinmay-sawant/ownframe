package replay

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// textRasterImages counts the cached rasters that hold an image, so a
// remembered decline does not read as a bake.
func textRasterImages() int {
	textRuns.mu.Lock()
	defer textRuns.mu.Unlock()

	count := 0
	for _, entry := range textRuns.entries {
		if entry.img != nil {
			count++
		}
	}

	return count
}

// checkTextRaster draws the text parity runs through the direct path and
// through the raster cache and requires identical pixels over an opaque
// backdrop. The second pass must hit every run, the two positions that differ
// by whole pixels must have baked one raster, and the run past the size cap
// must have baked none.
func checkTextRaster() error {
	ops, err := textParityOps()
	if err != nil {
		return err
	}

	face := textFace(&ops[0])
	if face == nil {
		return fmt.Errorf("the parity run has no face")
	}
	if first, last := phase(&ops[0], face), phase(&ops[4], face); first != last {
		return fmt.Errorf("parity positions %v and %v do not share a key", first, last)
	}

	direct := ebiten.NewImage(320, 240)
	cached := ebiten.NewImage(320, 240)
	defer direct.Dispose()
	defer cached.Dispose()

	before := textRasterImages()
	for pass := 0; pass < 2; pass++ {
		direct.Fill(color.White)
		cached.Fill(color.White)

		for i := range ops {
			drawTextDirect(&ops[i], direct)
			drawOp(cached, &ops[i], 0, 0)
		}

		a, b := make([]byte, 320*240*4), make([]byte, 320*240*4)
		direct.ReadPixels(a)
		cached.ReadPixels(b)
		if !bytes.Equal(a, b) {
			return fmt.Errorf("text raster pass %d differs in %d channels", pass, countDiff(a, b))
		}

		if got := textRasterImages() - before; got != 4 {
			return fmt.Errorf("pass %d baked %d rasters, want 4", pass, got)
		}
	}

	return nil
}
