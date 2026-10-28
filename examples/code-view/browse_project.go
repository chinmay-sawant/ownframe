package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// blockedProjectRoot reports whether root is a drive or filesystem root that
// must not open as a project. Drive roots scan too much and hide the real
// project name, so the picker redirects to a folder inside instead.
func blockedProjectRoot(root string) bool {
	trimmed := strings.TrimSpace(root)
	if trimmed == "" {
		return false
	}

	if trimmed == "/" {
		return true
	}

	// Windows drive roots like C:, C:\, or C:/.
	if len(trimmed) == 2 && isDriveLetter(trimmed[0]) && trimmed[1] == ':' {
		return true
	}

	if len(trimmed) == 3 && isDriveLetter(trimmed[0]) && trimmed[1] == ':' &&
		(trimmed[2] == '\\' || trimmed[2] == '/') {
		return true
	}

	// WSL drive mounts like /mnt/c.
	if len(trimmed) == len("/mnt/c") && strings.HasPrefix(trimmed, "/mnt/") &&
		isDriveLetter(trimmed[len(trimmed)-1]) {
		return true
	}

	return false
}

// isDriveLetter reports an ASCII drive letter.
func isDriveLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// browseStartRoot is the directory the folder browser opens on: the current
// project root when it is still on disk, else the home directory, else /.
func (a *App) browseStartRoot() string {
	if a.ws != nil && !a.ws.Sample && isDir(a.ws.Root) {
		return a.ws.Root
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" && isDir(home) {
		return home
	}

	return "/"
}

// startBrowse opens the folder browser on the last known directory.
func (a *App) startBrowse() {
	a.browseOpen = true
	a.browseRoot = a.browseStartRoot()
	a.fileAlert = ""
}

// browseMaxEntries caps the folder browser list. The browser shows immediate
// subdirectories only, so a deep root cannot flood the picker.
const browseMaxEntries = 100

// browseEntries lists up to browseMaxEntries immediate subdirectories of the
// root, sorted by name. Unreadable entries are skipped silently.
func (a *App) browseEntries() []recentRootRow {
	ents, err := os.ReadDir(a.browseRoot)
	if err != nil {
		return nil
	}

	rows := make([]recentRootRow, 0, len(ents))

	for _, ent := range ents {
		full := filepath.Join(a.browseRoot, ent.Name())
		if !isDir(full) {
			continue
		}

		rows = append(rows, recentRootRow{Path: full, Name: ent.Name()})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })

	if len(rows) > browseMaxEntries {
		rows = rows[:browseMaxEntries]
	}

	return rows
}

// browseDisplayName is the base name of the browser root, or the root
// itself when it has no base name (for example /).
func (a *App) browseDisplayName() string {
	if name := baseName(a.browseRoot); name != "" {
		return name
	}

	return a.browseRoot
}
