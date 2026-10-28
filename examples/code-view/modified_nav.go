package main

import (
	"sort"
	"strconv"
)

// modifiedNav is the floating widget's list of changed files: the paths in
// path order, the file the cursor points at, and whether the widget is open.
// Next and Prev walk the list, and Select jumps to one path.
type modifiedNav struct {
	Files []string
	Index int
	Open  bool
}

// newModifiedNav collects every changed file path once, sorted by path, and
// points at the first. Empty paths are skipped.
func newModifiedNav(changes changeSet) *modifiedNav {
	nav := &modifiedNav{Files: make([]string, 0, len(changes.Files))}
	seen := make(map[string]bool, len(changes.Files))

	for _, file := range changes.Files {
		if file.Path == "" || seen[file.Path] {
			continue
		}

		seen[file.Path] = true
		nav.Files = append(nav.Files, file.Path)
	}

	sort.Strings(nav.Files)

	return nav
}

// Len reports how many changed files the widget can visit.
func (n *modifiedNav) Len() int {
	return len(n.Files)
}

// Current returns the file at the cursor.
func (n *modifiedNav) Current() (string, bool) {
	if n.Index < 0 || n.Index >= len(n.Files) {
		return "", false
	}

	return n.Files[n.Index], true
}

// Select moves the cursor to path and reports whether the file is listed. A
// missing path leaves the cursor where it was.
func (n *modifiedNav) Select(path string) bool {
	for i, file := range n.Files {
		if file == path {
			n.Index = i

			return true
		}
	}

	return false
}

// Next moves to the following file and wraps at the end.
func (n *modifiedNav) Next() (string, bool) {
	if len(n.Files) == 0 {
		return "", false
	}

	n.Index = (n.Index + 1) % len(n.Files)

	return n.Files[n.Index], true
}

// Prev moves to the preceding file and wraps at the start.
func (n *modifiedNav) Prev() (string, bool) {
	if len(n.Files) == 0 {
		return "", false
	}

	n.Index = (n.Index - 1 + len(n.Files)) % len(n.Files)

	return n.Files[n.Index], true
}

// SetOpen shows or hides the floating widget.
func (n *modifiedNav) SetOpen(open bool) {
	n.Open = open
}

// Label names the cursor position for the widget, such as
// "3/7 backend/probe.go". It is empty when no file is listed.
func (n *modifiedNav) Label() string {
	path, ok := n.Current()
	if !ok {
		return ""
	}

	return strconv.Itoa(n.Index+1) + "/" + strconv.Itoa(len(n.Files)) + " " + path
}

// Keep narrows the list to paths and preserves the existing order. A file that
// is still listed keeps the cursor; when the selected file disappeared the
// cursor clamps into range.
func (n *modifiedNav) Keep(paths []string) {
	keep := make(map[string]bool, len(paths))

	for _, path := range paths {
		if path != "" {
			keep[path] = true
		}
	}

	current, had := n.Current()
	files := make([]string, 0, len(n.Files))

	for _, path := range n.Files {
		if keep[path] {
			files = append(files, path)
		}
	}

	n.Files = files

	if had {
		for i, path := range files {
			if path == current {
				n.Index = i

				return
			}
		}
	}

	n.Index = clampNavIndex(n.Index, len(files))
}

// clampNavIndex pins index inside a list of size files.
func clampNavIndex(index, files int) int {
	switch {
	case files <= 0:
		return 0
	case index < 0:
		return 0
	case index >= files:
		return files - 1
	default:
		return index
	}
}
