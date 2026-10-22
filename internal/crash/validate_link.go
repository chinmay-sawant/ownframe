package crash

import "path/filepath"

func resolveSymlinks(clean string) string {
	if r, err := filepath.EvalSymlinks(clean); err == nil {
		return r
	}
	if p, err := filepath.EvalSymlinks(filepath.Dir(clean)); err == nil {
		return filepath.Join(p, filepath.Base(clean))
	}
	return clean
}

func isSep(r rune) bool {
	return r == '/' || r == '\\'
}
