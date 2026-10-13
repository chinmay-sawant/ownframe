// Package styleinvalid tracks scoped CSS and layout invalidation.
//
// Instead of rebuilding every matched rule on a state change, callers record
// which box/property paths moved; the tracker exposes the dirty set so
// unmodified subtrees can be skipped during style and layout.
package styleinvalid

// Change is one touched style or layout dependency.
type Change struct {
	BoxID    string
	Property string
	Subtree  bool
}

// Tracker collects changes and reports dirty paths.
type Tracker struct {
	changes []Change
	// bound caps how many changes are accumulated before a full style pass.
	bound int
}

// New returns a tracker with the given coalescing bound.
func New(bound int) *Tracker {
	if bound <= 0 {
		bound = 1
	}
	return &Tracker{bound: bound}
}

// Touch records a change if it is not already present.
func (t *Tracker) Touch(c Change) {
	if t == nil {
		return
	}
	for _, existing := range t.changes {
		if existing == c {
			return
		}
	}
	t.changes = append(t.changes, c)
}

// Changes returns a copy of accumulated changes.
func (t *Tracker) Changes() []Change {
	if t == nil {
		return nil
	}
	out := make([]Change, len(t.changes))
	copy(out, t.changes)
	return out
}

// Full reports whether the dirty set exceeded the bound and a whole-document
// style pass is required.
func (t *Tracker) Full() bool {
	if t == nil {
		return true
	}
	return len(t.changes) > t.bound
}

// Reset clears accumulated changes after a pass.
func (t *Tracker) Reset() {
	if t == nil {
		return
	}
	t.changes = t.changes[:0]
}
