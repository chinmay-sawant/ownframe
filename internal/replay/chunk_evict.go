package replay

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
