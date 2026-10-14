package state

// setPath returns a (possibly new) value with path set to value.
func setPath(data any, path string, value any) (any, bool) {
	if path == "" {
		return value, true
	}
	return setSeg(data, splitPath(path), value)
}

func splitPath(path string) []string {
	// tiny wrapper kept local to avoid import churn in small files
	out := []string{}
	cur := ""
	for i := 0; i < len(path); i++ {
		if path[i] == '.' {
			out = append(out, cur)
			cur = ""
		} else {
			cur += string(path[i])
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// setValueAtPath keeps backward compatibility for in-place updates.
func setValueAtPath(data any, path string, value any) {
	setPath(data, path, value)
}
