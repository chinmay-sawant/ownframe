package styleinvalid

// StylePaths reports the unique style paths accumulated in this tracker.
func (t *Tracker) StylePaths() []StylePath {
	if t == nil {
		return nil
	}
	return StylePathsFromChanges(t.changes)
}

// Paths is an alias for StylePaths.
func (t *Tracker) Paths() []StylePath { return t.StylePaths() }

// TouchPath records a change for a style path when not already present.
func (t *Tracker) TouchPath(path StylePath) {
	if t == nil || path == "" {
		return
	}
	t.Touch(Change{Path: path})
}

// StyleDirty reports whether any accumulated change touches the given path.
func (t *Tracker) StyleDirty(path StylePath) bool {
	if t == nil {
		return false
	}
	for _, c := range t.changes {
		if c.StylePath() == path {
			return true
		}
	}
	return false
}
