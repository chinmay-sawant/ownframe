package state

// setPath returns a (possibly new) value with path set to value.
func setPath(data any, path string, value any) (any, bool) {
	if path == "" {
		return value, true
	}
	segs, ok := splitPath(path)
	if !ok {
		return data, false
	}
	return setSeg(data, segs, value)
}

// setValueAtPath keeps backward compatibility for in-place updates.
func setValueAtPath(data any, path string, value any) {
	setPath(data, path, value)
}
