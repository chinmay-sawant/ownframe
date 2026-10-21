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
// and stores the absolute form. An earlier value stays when rejected.
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

	// Reject .. before Clean: Clean resolves /allowed/../secret to
	// /secret, which would bypass a check on the cleaned path.
	for _, part := range strings.FieldsFunc(trimmed, isSep) {
		if part == ".." {
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

	for _, part := range strings.FieldsFunc(clean, isSep) {
		if part == ".." {
			return "", ErrBadDir
		}
	}

	if !filepath.IsAbs(clean) {
		abs, err := filepath.Abs(clean)
		if err != nil {
			return "", ErrBadDir
		}
		clean = abs
	}

	if !filepath.IsAbs(clean) {
		return "", ErrBadDir
	}

	return clean, nil
}

func isSep(r rune) bool {
	return r == '/' || r == '\\'
}
