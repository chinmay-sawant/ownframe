package replay

import (
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
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

var rasters rasterCache

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
		evicted.Dispose()
	}

	return img
}
