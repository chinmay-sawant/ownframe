package state

// StableIDs returns deterministic string IDs for repeated items. Items
// without an explicit key fall back to their index, so list diffs remain
// stable across reorderings only when callers supply keys.
func StableIDs(n int, key func(int) string) []string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		if key != nil {
			if k := key(i); k != "" {
				out = append(out, k)
				continue
			}
		}
		out = append(out, itoa(i))
	}
	return out
}

func itoa(i int) string {
	// Tiny, dependency-free int to string.
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
