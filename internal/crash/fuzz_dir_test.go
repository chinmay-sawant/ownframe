package crash_test

import (
	"strings"
	"sync"
	"testing"
	"unicode"
)

const fuzzLongDir = 8000

var fuzzDirMu sync.Mutex

func FuzzSetCrashDir(f *testing.F) {
	for _, seed := range []string{
		"", " ", "..", "../secret", "/allowed/../secret",
		"bad\x00dir", "bad\ndir", "bad\tdir", "a/../b",
		strings.Repeat("a", fuzzLongDir),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, dir string) {
		fuzzDirMu.Lock()
		defer fuzzDirMu.Unlock()
		fuzzSetDirOnce(t, dir)
	})
}

func mustRejectDir(dir string) bool {
	trimmed := strings.TrimSpace(dir)
	if trimmed == "" {
		return true
	}
	for _, r := range trimmed {
		if r == '\x00' || unicode.IsControl(r) {
			return true
		}
	}
	for _, part := range strings.FieldsFunc(trimmed, isPathSep) {
		if part == ".." {
			return true
		}
	}
	return false
}
