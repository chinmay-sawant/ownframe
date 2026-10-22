package replay

import (
	"bytes"
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// checkXformRepeat draws the page twice and requires bit-identical frames, so
// the cached raster blits stay deterministic across draws.
func checkXformRepeat(d *layout.Display) error {
	one, two := ebiten.NewImage(240, 140), ebiten.NewImage(240, 140)
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
		return fmt.Errorf("repeated draw differs in %d channels", countDiff(a, b))
	}

	return nil
}
