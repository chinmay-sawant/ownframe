package transfer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// trimWindows prefixes a Windows device name so "con.txt" stays usable.
func trimWindows(name string) string {
	base := name
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}

	if reserved[strings.ToLower(base)] {
		return "_" + name
	}

	return name
}

// clipName bounds a long name while keeping the extension.
func clipName(name string) string {
	const max = 120
	if len(name) <= max {
		return name
	}

	ext := filepath.Ext(name)
	if len(ext) > maxExtLen {
		ext = ""
	}

	return name[:max-len(ext)] + ext
}

// UniqueDestination picks a never-existing path under dir for raw. It tries
// raw, then "name (1)", "name (2)", and so on, comparing case-folded names
// so a Windows directory cannot produce two colliding entries. It creates
// dir when missing.
func UniqueDestination(dir, raw string) (string, error) {
	name := SafeName(raw)
	dir = filepath.Clean(dir)

	if err := os.MkdirAll(dir, transferDirPerm); err != nil {
		return "", err
	}

	for i := 0; i < 1000; i++ {
		candidate := filepath.Join(dir, Numbered(name, i))

		used, err := taken(dir, candidate)
		if err != nil {
			return "", err
		}

		if !used {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("transfer: no free destination name for %q", raw)
}

// Numbered inserts " (n)" before the extension. Zero keeps the name.
func Numbered(name string, n int) string {
	if n == 0 {
		return name
	}

	ext := filepath.Ext(name)

	return fmt.Sprintf("%s (%d)%s", name[:len(name)-len(ext)], n, ext)
}

// taken reports whether candidate or a case-folded twin exists in dir.
func taken(dir, candidate string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	want := strings.ToLower(filepath.Base(candidate))

	for _, entry := range entries {
		if strings.ToLower(entry.Name()) == want {
			return true, nil
		}
	}

	return false, nil
}
