package replay

// lookup returns the entry for key, if one has been baked or declined.
func (c *textRasterCache) lookup(key textRasterKey) (textRasterEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]

	return entry, ok
}

// raster returns the entry for key, running bake once while the key is new.
// bake reports whether the run can be cached at all; a decline is remembered
// as a zero entry.
func (c *textRasterCache) raster(key textRasterKey, bake func() (textRasterEntry, bool)) textRasterEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[textRasterKey]textRasterEntry)
	}
	if entry, ok := c.entries[key]; ok {
		return entry
	}

	entry, ok := bake()
	c.entries[key] = entry
	c.order = append(c.order, key)
	if ok {
		c.pixels += entry.pixels()
	}

	for len(c.order) > textRunMaxEntries || c.pixels > textRunMaxPixels {
		idx := -1
		var victim textRasterKey
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

		c.evict(victim)
		c.order = append(c.order[:idx], c.order[idx+1:]...)
	}

	return entry
}

// evict disposes one entry and forgets it. A declined entry has no image.
func (c *textRasterCache) evict(key textRasterKey) {
	entry, ok := c.entries[key]
	if !ok {
		return
	}

	c.pixels -= entry.pixels()
	if entry.img != nil {
		entry.img.Dispose()
	}
	delete(c.entries, key)
}

// flush drops every raster and every remembered decline. Only the benchmark
// needs a cold cache; a page reuses one for its whole life.
func (c *textRasterCache) flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, entry := range c.entries {
		if entry.img != nil {
			entry.img.Dispose()
		}
	}

	c.entries = nil
	c.order = nil
	c.pixels = 0
}
