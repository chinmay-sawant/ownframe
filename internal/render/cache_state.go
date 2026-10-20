package render

// Source returns the executed source the cached tree was parsed from.
func (c *Cache) Source() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.source
}

// StatePseudos reports which interactive pseudo-classes the cached sheets
// can match. The value is fixed for the cache's lifetime.
func (c *Cache) StatePseudos() StatePseudos {
	if c == nil {
		return allStatePseudos
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.pseudos
}
