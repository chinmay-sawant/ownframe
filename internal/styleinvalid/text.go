package styleinvalid

// TextRunKey identifies a shaped text run for reuse.
type TextRunKey struct {
	Face, Content string
	Color         uint32
	SubpixelX     int
}

// RunCache memoizes shaped text runs so repeated runs stay stable.
type RunCache map[TextRunKey]int

// NewRunCache returns an empty cache.
func NewRunCache() RunCache { return RunCache{} }

// Count returns how many distinct runs are cached.
func (c RunCache) Count() int {
	if c == nil {
		return 0
	}
	return len(c)
}

// Touch marks a run as seen, returning true if it was new.
func (c RunCache) Touch(k TextRunKey) bool {
	if c == nil {
		return false
	}
	if v, ok := c[k]; ok {
		c[k] = v + 1
		return false
	}
	c[k] = 1
	return true
}

// Seen reports whether a run has been touched.
func (c RunCache) Seen(k TextRunKey) bool {
	if c == nil {
		return false
	}
	_, ok := c[k]
	return ok
}

// Stable reports whether a run key is reusable.
func (c RunCache) Stable(k TextRunKey) bool {
	return TextRunStable(k)
}

// StableRuns reports distinct touched runs.
func (c RunCache) StableRuns() int {
	if c == nil {
		return 0
	}
	return len(c)
}

// Touches reports the total touch count.
func (c RunCache) Touches() int {
	if c == nil {
		return 0
	}
	n := 0
	for _, v := range c {
		n += v
	}
	return n
}
