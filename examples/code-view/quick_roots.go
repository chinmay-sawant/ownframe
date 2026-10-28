package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// quickRootRows lists well-known OS locations for the File drawer picker.
// The locations card stays user-dirs only: Home and Desktop on every OS,
// plus mounted volumes on macOS. Drive roots and /mnt mounts live in the
// folder browser card instead, where the Windows-vs-WSL choice is made.
// Every probe is best-effort: failures skip the entry and never error the page.
func (a *App) quickRootRows() []recentRootRow {
	var rows []recentRootRow
	seen := make(map[string]bool)
	add := func(path, name string) {
		if path == "" || name == "" || seen[path] {
			return
		}
		if len(rows) >= 10 {
			return
		}
		seen[path] = true
		rows = append(rows, recentRootRow{Path: path, Name: name})
	}

	home, homeErr := os.UserHomeDir()
	if homeErr != nil || home == "" {
		home = ""
	} else {
		add(home, "Home")
	}

	if home != "" {
		if d := filepath.Join(home, "Desktop"); isDir(d) {
			add(d, "Desktop")
		}
	}

	if runtime.GOOS == "darwin" {
		for _, r := range dirChildRows("/Volumes") {
			add(r.Path, r.Name)
		}
	}

	return rows
}

// isWSL reports a WSL2 Linux guest: /proc/version mentions microsoft
// (case-insensitive) or WSL_DISTRO_NAME is set.
func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	b, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(b)), "microsoft")
}

// dirChildRows lists each existing child directory of dir.
func dirChildRows(dir string) []recentRootRow {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []recentRootRow
	for _, e := range ents {
		full := filepath.Join(dir, e.Name())
		if !isDir(full) {
			continue
		}
		out = append(out, recentRootRow{Path: full, Name: e.Name()})
	}
	return out
}

func isDir(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}
