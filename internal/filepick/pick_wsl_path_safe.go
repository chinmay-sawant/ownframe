//go:build linux && !android

package filepick

import (
	"strings"
	"unicode"
)

// safeWindowsPath rejects empty paths, control characters, and Unicode line
// or paragraph separators before a Windows path is reused or converted.
func safeWindowsPath(picked string) bool {
	if picked == "" {
		return false
	}

	for _, r := range picked {
		if r == '\n' || r == '\r' || r == '\v' || r == '\f' ||
			r == '\u0085' || r == '\u2028' || r == '\u2029' ||
			unicode.IsControl(r) ||
			unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return false
		}
	}

	return true
}

func indexOfSep(s string, sep byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return i
		}
	}

	return -1
}

func hasDotSegment(path string) bool {
	for _, part := range strings.FieldsFunc(path, func(r rune) bool {
		return r == '/'
	}) {
		if part == "." || part == ".." {
			return true
		}
	}

	return false
}
