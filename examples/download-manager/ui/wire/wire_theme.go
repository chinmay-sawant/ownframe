package wire

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// themeState is the small settings file the example keeps beside the
// database. The store seam has no settings table, so the theme lives here.
type themeState struct {
	Dark bool `json:"dark"`
}

// themePath is the settings file, or "" in memory mode.
func (b *Backend) themePath() string {
	if b.dir == "" {
		return ""
	}

	return filepath.Join(b.dir, "ui.json")
}

// readDark loads the stored theme. A missing or broken file means light.
func readDark(path string) bool {
	if path == "" {
		return false
	}

	data, err := os.ReadFile(path) //nolint:gosec // reason: theme file is fixed ui.json in the own data dir.
	if err != nil {
		return false
	}

	var st themeState
	if err := json.Unmarshal(data, &st); err != nil {
		return false
	}

	return st.Dark
}

// writeDark stores the theme for the next launch.
func writeDark(path string, dark bool) error {
	if path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), wireDirPerm); err != nil {
		return err
	}

	data, err := json.Marshal(themeState{Dark: dark})
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, wireFilePerm)
}
