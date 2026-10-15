package replay

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// get returns the entry for key, baking it when the slot is missing or its
// signature or box moved. A changed signature with an unchanged box reuses
// the image and paints over it. The bake runs under the lock, as the raster
// caches do, so one slot never bakes twice.
func (c *chunkStore) get(key chunkKey, sig uint64, box image.Rectangle,
	bake func(*ebiten.Image)) *chunkEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	old := c.entries[key]
	if old != nil && old.sig == sig && old.box == box {
		return old
	}

	if c.entries == nil {
		c.entries = make(map[chunkKey]*chunkEntry)
	}

	var img *ebiten.Image
	if old != nil && old.box == box {
		old.img.Clear()
		img = old.img
	} else {
		img = ebiten.NewImage(box.Dx(), box.Dy())
		if old != nil {
			c.pixels -= old.img.Bounds().Dx() * old.img.Bounds().Dy()
			old.img.Dispose()
		}

		c.pixels += box.Dx() * box.Dy()
	}

	bake(img)

	if old == nil {
		c.order = append(c.order, key)
	}

	e := &chunkEntry{img: img, sig: sig, box: box}
	c.entries[key] = e
	c.evict(key)

	return e
}

// evict drops the oldest slots while the store is over either bound. protected
// is the entry just inserted or rebuilt and is never evicted here, so an entry
// that is itself larger than the pixel budget cannot be disposed and then
// returned to its caller.
func (c *chunkStore) evict(protected chunkKey) {
	for len(c.entries) > chunkMaxEntries || c.pixels > chunkMaxPixels {
		victim := chunkKey{}
		idx := -1
		for i, candidate := range c.order {
			if candidate == protected {
				continue
			}
			victim = candidate
			idx = i
			break
		}
		if idx < 0 {
			// Only the protected entry remains; let it stay.
			break
		}

		c.order = append(c.order[:idx], c.order[idx+1:]...)

		if e := c.entries[victim]; e != nil {
			delete(c.entries, victim)
			c.pixels -= e.img.Bounds().Dx() * e.img.Bounds().Dy()
			e.img.Dispose()
		}
	}
}
