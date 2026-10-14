package styleinvalid

// StylePath is a string path (box/property) that style invalidation can target.
type StylePath = string

// StylePath returns the combined box/property path for this change.
func (c Change) StylePath() StylePath {
	if c.Path != "" {
		return c.Path
	}
	if c.Property == "" {
		return c.BoxID
	}
	if c.BoxID == "" {
		return c.Property
	}
	return c.BoxID + "/" + c.Property
}

// StylePathsFromChanges collects unique style paths touched by changes.
func StylePathsFromChanges(changes []Change) []StylePath {
	seen := map[StylePath]bool{}
	var out []StylePath
	for _, c := range changes {
		p := c.StylePath()
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

// StylePaths is an alias for StylePathsFromChanges.
func StylePaths(changes []Change) []StylePath { return StylePathsFromChanges(changes) }

// InvalidateStylePaths converts touched paths into a change list. Existing
// changes with matching paths are preserved; new ones are appended.
func InvalidateStylePaths(paths []StylePath) []Change {
	out := make([]Change, 0, len(paths))
	seen := map[StylePath]bool{}
	for _, p := range paths {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, Change{Path: p})
	}
	return out
}

// InvalidateStylePath builds one change from a path.
func InvalidateStylePath(path StylePath) Change { return Change{Path: path} }
