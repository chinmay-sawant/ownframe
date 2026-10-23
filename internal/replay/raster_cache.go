package replay

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
	"math"
	"sync"
)

// rasterMaxEntries and rasterMaxTotalPixels bound what one page caches. A
// page with more shapes than this keeps the newest; the oldest rasters are
// disposed.
const (
	rasterMaxEntries     = 256
	rasterMaxTotalPixels = 4 << 20
)

// rasterCache keeps baked vector rasters, oldest insertion first.
type rasterCache struct {
	mu      sync.Mutex
	entries map[rasterKey]*ebiten.Image
	order   []rasterKey
	pixels  int
}

var defaultRenderer Renderer

// raster returns the raster for key, baking it into an image of size once.
// The bake draws the shape so the origin's integer part is the image's
// top-left, which blitRaster then restores.
func (c *rasterCache) raster(key rasterKey, size image.Point, bake func(*ebiten.Image)) *ebiten.Image {
	c.mu.Lock()
	defer c.mu.Unlock()

	if img := c.entries[key]; img != nil {
		return img
	}

	if c.entries == nil {
		c.entries = make(map[rasterKey]*ebiten.Image)
	}

	img := ebiten.NewImage(size.X, size.Y)
	bake(img)
	c.entries[key] = img
	c.order = append(c.order, key)
	c.pixels += size.X * size.Y

	for len(c.order) > rasterMaxEntries || c.pixels > rasterMaxTotalPixels {
		victim := rasterKey{}
		idx := -1
		for i, candidate := range c.order {
			if candidate == key {
				continue
			}
			victim = candidate
			idx = i
			break
		}
		if idx < 0 {
			// Only the entry being returned remains; keep it.
			break
		}

		c.order = append(c.order[:idx], c.order[idx+1:]...)
		evicted := c.entries[victim]
		delete(c.entries, victim)
		c.pixels -= evicted.Bounds().Dx() * evicted.Bounds().Dy()
		evicted.Deallocate()
	}

	return img
}

// rasterKey names one cached vector rasterization. It holds everything that
// shapes the pixels: the variant, the size and corner radii, the stroke width
// and mask, the color, and the sub-pixel offset of the op's origin. The
// integer part of the position is not in the key, because the raster bakes at
// the fractional offset and blits at the integer one, so equal shapes at
// different integer positions share one raster.
type rasterKey struct {
	variant uint8
	mask    uint8
	col     color.RGBA
	w, h    float32
	width   float32
	r, ry   [4]float32
	fx, fy  float32
}

// Raster variants.
const (
	rasterFill uint8 = iota + 1
	rasterStroke
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
