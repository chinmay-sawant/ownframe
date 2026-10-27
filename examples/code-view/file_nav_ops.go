package main

import "strings"

type changeNavigation struct {
	path   string
	mode   string
	group  int
	row    int
	scroll int
	valid  bool
}

// fileNavChangeStep visits change blocks in the selected file.
func (a *App) fileNavChangeStep(delta int) {
	if a.fileNav == nil {
		return
	}
	path, ok := a.fileNav.Current()
	if !ok {
		return
	}
	if a.mode == "changes" {
		if a.diffPath != path || a.diffSource != "working" {
			a.openChangedPath(path)
		}
	} else if doc := a.activeDoc(); doc == nil || doc.Path != path {
		a.openChangedPath(path)
	}

	rows := a.fileNavChangeRows(path)
	if len(rows) == 0 {
		return
	}
	index := a.fileNavChangeIndex(path, rows) + delta
	if index < 0 || index >= len(rows) {
		return
	}
	var scroll int
	if a.split && a.mode != "changes" {
		a.scrollGroupTo(a.activeGroup(), rows[index]*rowH)
		scroll = a.scrollGroup[a.activeGroup()]
	} else {
		a.scrollTo(rows[index] * rowH)
		scroll = a.scrollY
	}
	a.changeNav = changeNavigation{path: path, mode: a.mode, group: a.activeGroup(), row: rows[index], scroll: scroll, valid: true}
	a.refresh()
}

func (a *App) fileNavChangeIndex(path string, rows []int) int {
	scroll := a.scrollY
	if a.split && a.mode != "changes" {
		scroll = a.scrollGroup[a.activeGroup()]
	}
	current := scroll / rowH
	if a.changeNav.valid && a.changeNav.path == path &&
		a.changeNav.mode == a.mode && a.changeNav.group == a.activeGroup() && a.changeNav.scroll == scroll {
		current = a.changeNav.row
	}
	index := -1
	for i, row := range rows {
		if row > current {
			break
		}
		index = i
	}
	if index < 0 && len(rows) > 0 {
		return 0
	}
	return index
}

func (a *App) fileNavChangeRows(path string) []int {
	var rows []int
	if a.mode == "changes" {
		if a.diffPath != path || a.diffSource != "working" {
			return nil
		}
		return diffChangeStarts(a.diffDisplayRows())
	}
	doc := a.activeDoc()
	if doc == nil || doc.Path != path {
		return nil
	}
	previous := -1
	for _, line := range inlineChangeLines(a.changedLinesFor(path)) {
		if line != previous+1 {
			row := line - 1
			if a.wrap {
				row = doc.wrapped(a.codeColsForGroup(a.activeGroup())).firstRow(row)
			}
			rows = append(rows, row)
		}
		previous = line
	}
	return rows
}

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

	nav.Open = !a.paletteOpen && !a.searchOpen && !a.branchOpen
	nav.Name = baseName(path)
	nav.Dir = dirName(path)
	nav.Label = nav.Name
	nav.Path = nav.Dir
	nav.Position = a.fileNav.Index + 1
	nav.Total = a.fileNav.Len()
	nav.Add = file.Add
	nav.Del = file.Del
	nav.Multiple = nav.Total > 1
	rows := a.fileNavChangeRows(path)
	index := a.fileNavChangeIndex(path, rows)
	nav.CanPrevChange = index > 0
	nav.CanNextChange = index+1 < len(rows)

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
