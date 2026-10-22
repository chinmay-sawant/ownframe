package studio

import (
	"errors"
	"os"
	"path/filepath"
)

// dirPerm and filePerm keep the theme tree owner-only (gosec G301/G306).
const (
	dirPerm  = 0o750
	filePerm = 0o600
)

// themePath is os.UserConfigDir()/ownframe/theme-studio/theme.css, so on
// Linux ~/.config/ownframe/theme-studio/theme.css. It makes the directory
// so a first save needs no other setup.
func themePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "ownframe", "theme-studio")
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return "", err
	}

	return filepath.Join(dir, "theme.css"), nil
}

// loadTheme reads the theme file. A missing file is an empty sheet, not an
// error, so a first run starts from the default and seeds the file.
func loadTheme(path string) (string, error) {
	css, err := os.ReadFile(path) //nolint:gosec // reason: path comes from themePath under the user config dir or a test temp dir, not from page input.
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return string(css), nil
}

// saveTheme writes the sheet, making the parent directory first because the
// page keeps watching the path after the file is gone.
func saveTheme(path, css string) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(css), filePerm)
}
