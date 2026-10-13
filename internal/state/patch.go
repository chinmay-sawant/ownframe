package state

// Patch describes a small state update. Changed lists the property paths that
// moved; Values holds the projected slice for those paths. A nil or empty
// Changed set means "whole state replaced", which is what SetData produces.
type Patch struct {
	Origin  Origin
	Changed []string
	Values  map[string]any
}

// IsWhole reports whether the patch replaces the entire state.
func (p Patch) IsWhole() bool {
	return len(p.Changed) == 0
}

// Has reports whether path is among the changed properties.
func (p Patch) Has(path string) bool {
	for _, c := range p.Changed {
		if c == path {
			return true
		}
	}
	return false
}
