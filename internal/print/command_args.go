package print

import (
	"errors"
	"strings"
	"unicode"
)

// ErrBadPath means the print path contains a newline, control character,
// Unicode line separator, or unsafe dot segment and was rejected before
// any command was invoked.
var ErrBadPath = errors.New("ownframe: print path contains a control character")

// badPath reports whether path contains a newline, another control character,
// a Unicode line separator, or a "." or ".." path segment.
func badPath(path string) bool {
	if path == "" {
		return true
	}

	for _, r := range path {
		if r == '\n' || r == '\r' || r == '\v' || r == '\f' ||
			r == '\u0085' || r == '\u2028' || r == '\u2029' ||
			unicode.IsControl(r) ||
			unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return true
		}
	}

	for _, part := range strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	}) {
		if part == "." || part == ".." {
			return true
		}
	}

	return false
}

func ValidPath(path string) error {
	if badPath(path) {
		return ErrBadPath
	}

	return nil
}

func validPath(path string) error {
	return ValidPath(path)
}

func lpFileArg(path string) []string {
	return []string{"--", path}
}

func openFileArg(path string) []string {
	return []string{"--", path}
}

func appleScriptPrint(path string) []string {
	return []string{"-e", `on run argv`, "-e", `tell application "Preview" to print (POSIX file (item 1 of argv))`, "-e", `end run`, "--", path}
}

func powershellPrint(path string) []string {
	quoted := strings.ReplaceAll(path, "'", "''")

	return []string{
		"-NoProfile",
		"-Command",
		"Start-Process -FilePath '" + quoted + "' -Verb Print",
	}
}
