package print

import (
	"errors"
	"path/filepath"
	"strings"
	"unicode"
)

// ErrBadPath means the print path was empty, held a control
// character, or escaped after cleaning and was rejected.
var ErrBadPath = errors.New("ownframe: print path contains a control character")

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
	clean := filepath.Clean(path)
	if clean == "." {
		return true
	}
	for _, part := range strings.FieldsFunc(clean, isSep) {
		if part == ".." {
			return true
		}
	}
	return false
}

func canonicalPath(path string) (string, error) {
	if badPath(path) {
		return "", ErrBadPath
	}
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		abs, err := filepath.Abs(clean)
		if err != nil {
			return "", ErrBadPath
		}
		clean = abs
	}
	return filepath.Clean(clean), nil
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

func isSep(r rune) bool {
	return r == '/' || r == '\\'
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
