package replay

// touch marks key most-recently used by moving it to the back of the order,
// so a later evict drops the least-recently used slot first. Every get path
// calls it: a hit, a signature rebuild, and a box rebuild all mean the chunk
// painted this frame, and the next paint only needs the chunks of the region
// it draws.
func (c *chunkStore) touch(key chunkKey) {
	for i, candidate := range c.order {
		if candidate != key {
			continue
		}
		c.order = append(c.order[:i], c.order[i+1:]...)
		c.order = append(c.order, key)
		return
	}
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
			e.img.Deallocate()
		}
	}
}
