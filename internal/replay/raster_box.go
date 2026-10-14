package replay

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// rasterPad is the integer slack per side that covers antialiasing spill.
const rasterPad = 2

// rasterMaxSide and rasterMaxPixels bound one cached raster. Past them the
// caller draws the shape directly.
const (
	rasterMaxSide   = 512
	rasterMaxPixels = 1 << 20
)

// rasterBox returns the integer blit origin and the image size for a shape
// whose filled box is x,y,w,h, grown by grow on every side. ok is false when
// the shape cannot be cached.
func rasterBox(x, y, w, h, grow float32) (left, top int, size image.Point, ok bool) {
	// Float32 cannot hold a small pixel offset next to a huge origin, so a
	// shape far from the page origin keeps the direct path.
	if x < -1e6 || y < -1e6 || x > 1e6 || y > 1e6 {
		return 0, 0, image.Point{}, false
	}

	left = int(math.Floor(float64(x-grow))) - rasterPad
	top = int(math.Floor(float64(y-grow))) - rasterPad
	right := int(math.Ceil(float64(x+w+grow))) + rasterPad
	bottom := int(math.Ceil(float64(y+h+grow))) + rasterPad
	size = image.Pt(right-left, bottom-top)
	if size.X <= 0 || size.Y <= 0 ||
		size.X > rasterMaxSide || size.Y > rasterMaxSide ||
		size.X*size.Y > rasterMaxPixels {
		return 0, 0, image.Point{}, false
	}

	return left, top, size, true
}

// blitRaster copies a baked raster to dst at an integer origin. An integer
// translation leaves every texel on its center, so the copy is exact.
func blitRaster(dst, img *ebiten.Image, left, top int) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(left), float64(top))
	dst.DrawImage(img, opts)
}
