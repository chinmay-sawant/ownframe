package state

// Tracker keeps an ordered log of state patches so callers can ask what
// changed without re-running a render.
type Tracker struct {
	patches []Patch
}

// NewTracker returns an empty tracker.
func NewTracker() *Tracker { return &Tracker{} }

// Record appends a patch.
func (t *Tracker) Record(p Patch) {
	if t == nil {
		return
	}
	t.patches = append(t.patches, p)
}

// Last returns the most recent patch and ok.
func (t *Tracker) Last() (Patch, bool) {
	if t == nil || len(t.patches) == 0 {
		return Patch{}, false
	}
	return t.patches[len(t.patches)-1], true
}

// All returns a copy of every recorded patch.
func (t *Tracker) All() []Patch {
	if t == nil {
		return nil
	}
	out := make([]Patch, len(t.patches))
	copy(out, t.patches)
	return out
}

// Reset clears the log.
func (t *Tracker) Reset() {
	if t == nil {
		return
	}
	t.patches = t.patches[:0]
}
