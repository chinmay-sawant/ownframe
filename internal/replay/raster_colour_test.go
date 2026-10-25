package replay

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// TestRasterCacheReusesShapeAcrossColours pins the white-mask bake: a shape
// whose colour changes reuses one cached raster and applies the colour at
// blit time, so a colour animation does not re-rasterize its path per frame.
func TestRasterCacheReusesShapeAcrossColours(t *testing.T) {
	var renderer Renderer
	defer renderer.Close()

	dst := ebiten.NewImage(40, 40)
	defer dst.Deallocate()

	op := &layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 20, H: 20,
		R: 1, G: 0, B: 0, Alpha: 1, Radius: 4,
	}
	renderer.fillRect(dst, op, 0, 0)
	if got := len(renderer.rasters.entries); got != 1 {
		t.Fatalf("rasters = %d, want one", got)
	}

	op.R, op.G, op.B = 0, 0, 1
	renderer.fillRect(dst, op, 0, 0)
	if got := len(renderer.rasters.entries); got != 1 {
		t.Fatalf("a colour change baked a second raster: %d entries", got)
	}
}
