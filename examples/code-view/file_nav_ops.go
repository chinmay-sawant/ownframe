package main

import "strings"

// syncFileNav rebuilds the changed-file navigator and keeps the current path
// selected when it still exists.
func (a *App) syncFileNav() {
	path := ""
	if a.fileNav != nil {
		path, _ = a.fileNav.Current()
	}

	a.fileNav = newModifiedNav(a.changes)
	if path != "" {
		a.fileNav.Select(path)
	}

	a.summary = summarizeChanges(a.changes)
}

// fileNavViewState prepares the bottom-center widget.
func (a *App) fileNavViewState() fileNavView {
	nav := fileNavView{}
	if a.fileNav == nil || a.fileNav.Len() == 0 {
		return nav
	}

	path, ok := a.fileNav.Current()
	if !ok {
		return nav
	}

	file := findFile(a.changes.Files, path)
	if file == nil {
		return nav
	}

	nav.Open = !a.paletteOpen && !a.searchOpen && !a.branchOpen && !a.editing
	nav.Name = baseName(path)
	nav.Dir = dirName(path)
	nav.Label = nav.Name
	nav.Path = nav.Dir
	nav.Position = a.fileNav.Index + 1
	nav.Total = a.fileNav.Len()
	nav.Add = file.Add
	nav.Del = file.Del
	nav.Multiple = nav.Total > 1

	return nav
}

// fileNavStep moves the changed-file cursor and opens the new file or diff.
func (a *App) fileNavStep(delta int) {
	if a.fileNav == nil || a.fileNav.Len() == 0 {
		return
	}

	var (
		path string
		ok   bool
	)

	if delta < 0 {
		path, ok = a.fileNav.Prev()
	} else {
		path, ok = a.fileNav.Next()
	}

	if !ok {
		return
	}

	a.openChangedPath(path)
}

// openChangedPath opens one changed file in the current mode.
func (a *App) openChangedPath(path string) {
	if path == "" {
		return
	}

	if a.mode == "changes" {
		a.selectDiff(path)
	} else {
		a.openFile(path)
	}

	if a.fileNav != nil {
		a.fileNav.Select(path)
	}

	a.refresh()
}

// fileNavOpen opens the current changed file.
func (a *App) fileNavOpen() {
	if a.fileNav == nil {
		return
	}

	path, ok := a.fileNav.Current()
	if !ok {
		return
	}

	a.openChangedPath(path)
}

// fileNavSummary is the status-bar sentence for the changed set.
func (a *App) fileNavSummary() string {
	if a.summary.Total == 0 {
		return ""
	}

	return strings.TrimSpace(a.summary.Label())
}
