package state

import "strings"

func splitPath(path string) ([]string, bool) {
	if path == "" {
		return nil, true
	}
	segs := strings.Split(path, ".")
	for _, s := range segs {
		if s == "" {
			return nil, false
		}
	}
	return segs, true
}
