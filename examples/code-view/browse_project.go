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
	a.browseList.reset()
	a.fileAlert = ""
}

// browseMaxEntries caps the folder browser list. The browser shows immediate
// subdirectories only, so a deep root cannot flood the picker.
const browseMaxEntries = 100

// browseEntries lists up to browseMaxEntries immediate subdirectories of the
// root, sorted by name. Unreadable entries are skipped silently. An empty
// root is the virtual drives level (This PC) and lists every existing
// Windows drive instead of reading a directory.
func (a *App) browseEntries() []recentRootRow {
	if a.browseRoot == "" {
		return availableDrives(isDir)
	}

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
// itself when it has no base name (for example /). The virtual drives
// level (empty root) shows as This PC.
func (a *App) browseDisplayName() string {
	if a.browseRoot == "" {
		return "This PC"
	}

	if name := baseName(a.browseRoot); name != "" {
		return name
	}

	return a.browseRoot
}

// isBrowseDriveRoot reports a Windows drive root such as C:, C:\, or C:/.
// Only drive roots step out to the virtual drives level on Up; the Unix
// root / keeps its existing behavior and never becomes drives.
func isBrowseDriveRoot(root string) bool {
	if len(root) == 2 && isDriveLetter(root[0]) && root[1] == ':' {
		return true
	}

	if len(root) == 3 && isDriveLetter(root[0]) && root[1] == ':' &&
		(root[2] == '\\' || root[2] == '/') {
		return true
	}

	return false
}

// isWindowsDrivePath reports a path on a Windows drive (D:\, D:/, D:foo).
// The check is textual so drive logic stays testable on every OS.
func isWindowsDrivePath(path string) bool {
	return len(path) >= 2 && isDriveLetter(path[0]) && path[1] == ':'
}

// driveSepIndex is the separator position in a drive-rooted path: D:\x
// carries its first inner separator at index 2, and anything at or below
// that index is the root itself.
const driveSepIndex = 2

// browseParentDir returns the parent of root. Windows drive paths compare
// both separators explicitly so Up behaves the same on every OS; anything
// else follows the host filepath rules.
func browseParentDir(root string) string {
	if isWindowsDrivePath(root) {
		if isBrowseDriveRoot(root) {
			return root
		}

		end := len(root)
		for end > 3 && (root[end-1] == '/' || root[end-1] == '\\') {
			end--
		}

		trimmed := root[:end]
		for i := len(trimmed) - 1; i >= 2; i-- {
			if trimmed[i] == '/' || trimmed[i] == '\\' {
				if i == driveSepIndex {
					return trimmed[:2] + `\`
				}

				return trimmed[:i]
			}
		}

		return root[:2]
	}

	return filepath.Dir(root)
}

// browseUpTarget resolves one Up step from root. ok is false when Up goes
// nowhere: the drives level itself, or the Unix root /. A drive root steps
// out to the virtual drives level, reported as the empty sentinel.
func browseUpTarget(root string) (next string, ok bool) {
	if root == "" {
		return "", false
	}

	if isBrowseDriveRoot(root) {
		return "", true
	}

	parent := browseParentDir(root)
	if parent == "" || parent == root {
		return root, false
	}

	return parent, true
}

// driveCapHint sizes the drive row slice for the common few-drive case;
// availableDrives grows it when a machine has more.
const driveCapHint = 4

// availableDrives probes A:\ through Z:\ and returns one row per existing
// drive. exists is isDir in the browser and a stub in tests, so the level
// stays deterministic on machines without drive letters. The 26 letters
// bound the rows; browseMaxEntries caps them further.
func availableDrives(exists func(string) bool) []recentRootRow {
	rows := make([]recentRootRow, 0, driveCapHint)

	for c := byte('A'); c <= byte('Z'); c++ {
		drive := string([]byte{c, ':', '\\'})
		if !exists(drive) {
			continue
		}

		rows = append(rows, recentRootRow{Path: drive, Name: drive})

		if len(rows) >= browseMaxEntries {
			break
		}
	}

	return rows
}
