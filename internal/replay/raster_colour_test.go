package replay

import (
	"image"
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

// TestRasterCacheTouchOnHit verifies that a cache hit moves the key to the back
// of the eviction order for true LRU retention.
func TestRasterCacheTouchOnHit(t *testing.T) {
	var c rasterCache
	k1 := rasterKey{variant: 1, w: 10, h: 10}
	k2 := rasterKey{variant: 1, w: 20, h: 20}

	c.raster(k1, image.Pt(10, 10), func(img *ebiten.Image) {})
	c.raster(k2, image.Pt(20, 20), func(img *ebiten.Image) {})

	if len(c.order) != 2 || c.order[0] != k1 || c.order[1] != k2 {
		t.Fatalf("order = %v, want [%v, %v]", c.order, k1, k2)
	}

	// Hit k1: must move to the back of order.
	c.raster(k1, image.Pt(10, 10), func(img *ebiten.Image) {})
	if len(c.order) != 2 || c.order[0] != k2 || c.order[1] != k1 {
		t.Fatalf("order after hit = %v, want [%v, %v]", c.order, k2, k1)
	}
}
