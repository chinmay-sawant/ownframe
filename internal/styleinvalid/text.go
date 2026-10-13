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
func (c RunCache) Count() int { return len(c) }

// Touch marks a run as seen, returning true if it was new.
func (c RunCache) Touch(k TextRunKey) bool {
	if _, ok := c[k]; ok {
		return false
	}
	c[k] = 1
	return true
}
