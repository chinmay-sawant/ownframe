package crash

import (
	"errors"
	"path/filepath"
	"strings"
	"unicode"
)

// ErrEmptyDir reports that SetDir was given an empty path.
var ErrEmptyDir = errors.New("crash: empty directory")

// ErrBadDir reports that SetDir was given an invalid path.
var ErrBadDir = errors.New("crash: invalid directory")

// SetDir sets the folder Write uses. It rejects an empty or unsafe path
// and stores a cleaned path. An earlier configured value stays in
// place when the new value is rejected.
func SetDir(dir string) error {
	clean, err := validateDir(dir)
	if err != nil {
		return err
	}

	mu.Lock()
	custom = clean
	mu.Unlock()
	return nil
}

func validateDir(dir string) (string, error) {
	trimmed := strings.TrimSpace(dir)
	if trimmed == "" {
		return "", ErrEmptyDir
	}

	for _, r := range trimmed {
		if r == '\x00' || r == '\n' || r == '\r' || r == '\v' ||
			r == '\f' || r == '\u0085' || r == '\u2028' ||
			r == '\u2029' || unicode.IsControl(r) ||
			unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return "", ErrBadDir
		}
	}

	clean := filepath.Clean(trimmed)
	if clean == "." || clean == string(filepath.Separator) {
		return "", ErrBadDir
	}

	if len(clean) == 2 && clean[1] == ':' {
		return "", ErrBadDir
	}

	separator := string(filepath.Separator)
	for _, part := range strings.Split(trimmed, separator) {
		if part == "." || part == ".." {
			return "", ErrBadDir
		}
	}
	for _, part := range strings.Split(trimmed, "\\") {
		if part == "." || part == ".." {
			return "", ErrBadDir
		}
	}

	if !filepath.IsAbs(clean) {
		if clean == ".." || strings.HasPrefix(clean, ".."+separator) {
			return "", ErrBadDir
		}
	}

	return clean, nil
}
