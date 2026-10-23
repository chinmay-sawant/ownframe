package replay

import (
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// imageCacheLimit bounds entries kept. maxImageCachePixels bounds total
// pixels, so many large images evict oldest first.
const imageCacheLimit = 64
const maxImageCachePixels = 64 << 20

// imageCache keeps decoded images by payload key, oldest first. Hits
// resolve under a read lock; decode runs lock-free and only the GPU upload
// plus insert hold the write lock, with a second hit check.
type imageCache struct {
	mu      sync.RWMutex
	entries map[uint64]*ebiten.Image
	order   []uint64
	sizes   map[uint64]int
	total   int
}

func (c *imageCache) get(key uint64) *ebiten.Image {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entries[key]
}

func (c *imageCache) put(key uint64, src image.Image, px int) *ebiten.Image {
	c.mu.Lock()
	defer c.mu.Unlock()
	if img := c.entries[key]; img != nil {
		return img
	}
	img := ebiten.NewImageFromImage(src)
	if c.entries == nil {
		c.entries = make(map[uint64]*ebiten.Image)
		c.sizes = make(map[uint64]int)
	}
	c.entries[key] = img
	c.order = append(c.order, key)
	c.sizes[key] = px
	c.total += px
	for len(c.order) > imageCacheLimit || c.total > maxImageCachePixels {
		old := c.order[0]
		c.order = c.order[1:]
		if e := c.entries[old]; e != nil {
			e.Deallocate()
		}
		delete(c.entries, old)
		c.total -= c.sizes[old]
		delete(c.sizes, old)
		if len(c.order) == 0 {
			break
		}
	}
	return img
}
