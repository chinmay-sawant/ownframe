package crash

import (
	"errors"
	"path/filepath"
	"strings"
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

	clean := filepath.Clean(trimmed)
	if clean == "." || clean == string(filepath.Separator) {
		return "", ErrBadDir
	}

	if !filepath.IsAbs(clean) {
		if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return "", ErrBadDir
		}
	}

	return clean, nil
}
