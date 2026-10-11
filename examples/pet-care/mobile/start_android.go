//go:build android

package mobile

import "path/filepath"

func Start(configDir string) string {
	if configDir == "" {
		return "pet care: Android app data directory is empty"
	}
	path := filepath.Join(configDir, "ownframe", "pet-care", "pet-care.db")
	if err := start(path); err != nil {
		return err.Error()
	}
	return ""
}
