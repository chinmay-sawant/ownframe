package replay

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// get returns the entry for key, baking it when the slot is missing or its
// signature or box moved. A changed signature with an unchanged box reuses
// the image and paints over it. The bake stays under the lock, as measured
// by BenchmarkChunkCardDamage/chunked: bake draws into a slot image a
// concurrent bake could share, and Ebiten calls belong on the draw path,
// so moving it out races. The replay suite stays green under -race with
// the bake inline. Fewer misses, not a split lock, cuts stall time.
func (c *chunkStore) get(key chunkKey, sig uint64, box image.Rectangle,
	bake func(*ebiten.Image)) *chunkEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	old := c.entries[key]
	if old != nil && old.sig == sig && old.box == box {
		c.touch(key)
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
			old.img.Deallocate()
		}

		c.pixels += box.Dx() * box.Dy()
	}

	bake(img)

	if old == nil {
		c.order = append(c.order, key)
	} else {
		c.touch(key)
	}

	e := &chunkEntry{img: img, sig: sig, box: box}
	c.entries[key] = e
	c.evict(key)

	return e
}
