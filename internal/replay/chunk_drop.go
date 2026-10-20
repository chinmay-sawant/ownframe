package replay

import "github.com/chinmay-sawant/blinkless/layout"

// DropStaleChunks removes the chunk entries that belong to a display the
// page replaced. A chunk key holds its display, so a stale entry keeps a
// whole display list, with its text and boxes, alive until the entry count
// evicts it.
func DropStaleChunks(keep *layout.Display) {
	chunks.dropStale(keep)
}

// dropStale frees the entries whose display is not the current one.
func (c *chunkStore) dropStale(keep *layout.Display) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) == 0 {
		return
	}

	order := c.order[:0]
	for _, key := range c.order {
		entry, ok := c.entries[key]
		if !ok {
			continue
		}
		if key.d != keep {
			c.pixels -= entry.img.Bounds().Dx() * entry.img.Bounds().Dy()
			entry.img.Dispose()
			delete(c.entries, key)

			continue
		}
		order = append(order, key)
	}
	c.order = order
}
