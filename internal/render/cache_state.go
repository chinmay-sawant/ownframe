package render

// Source returns the executed source the cached tree was parsed from.
func (c *Cache) Source() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.source
}

// StatePseudos reports which interactive pseudo-classes the cached sheets
// can match. The value follows the styled document, so a patch that rewrites
// a style element reports the states the new sheets can match.
func (c *Cache) StatePseudos() StatePseudos {
	if c == nil {
		return allStatePseudos
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.styled == nil {
		return allStatePseudos
	}

	return statePseudosOf(c.styled)
}
