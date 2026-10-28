package main

import (
	"context"
	"strings"
)

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
	nav := a.currentFileNav()
	if nav == nil {
		return
	}
	path, ok := nav.Current()
	if !ok {
		return
	}
	if a.mode == "changes" {
		if a.diffSource == "branch" && a.branchPath != path {
			a.openChangedPath(path)
		}
		if (a.diffSource == "commit" && a.commitFile != path) || (a.diffSource == "working" && a.diffPath != path) {
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
		if (a.diffSource == "commit" && a.commitFile != path) || (a.diffSource == "branch" && a.branchPath != path) || (a.diffSource == "working" && a.diffPath != path) {
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

func (a *App) currentFileNav() *modifiedNav {
	if a.mode == "changes" && a.diffSource == "branch" {
		nav := newModifiedNav(a.branchChanges)
		nav.Select(a.branchPath)
		return nav
	}
	if a.mode != "changes" || a.diffSource != "commit" {
		return a.fileNav
	}
	commit := a.findCommit(a.commitOpen)
	if commit == nil {
		return nil
	}
	changes := changeSet{}
	for _, file := range commit.Files {
		changes.Files = append(changes.Files, changedFile{Path: file.Path})
	}
	nav := newModifiedNav(changes)
	nav.Select(a.commitFile)
	return nav
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
	files := a.currentFileNav()
	if files == nil || files.Len() == 0 {
		return nav
	}

	path, ok := files.Current()
	if !ok {
		return nav
	}

	file := findFile(a.changes.Files, path)
	if a.mode == "changes" && a.diffSource == "branch" {
		file = findFile(a.branchChanges.Files, path)
	}
	if a.mode == "changes" && a.diffSource == "commit" {
		file = nil
		if commit := a.findCommit(a.commitOpen); commit != nil {
			for _, entry := range commit.Files {
				if entry.Path == path {
					file = &changedFile{Path: entry.Path, Add: entry.Add, Del: entry.Del}
					break
				}
			}
		}
	}
	if file == nil {
		return nav
	}

	nav.Open = !a.paletteOpen && !a.searchOpen && !a.branchOpen
	nav.Name = baseName(path)
	nav.Dir = dirName(path)
	nav.Label = nav.Name
	nav.Path = nav.Dir
	nav.Position = files.Index + 1
	nav.Total = files.Len()
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
	nav := a.currentFileNav()
	if nav == nil || nav.Len() == 0 {
		return
	}

	var (
		path string
		ok   bool
	)

	if delta < 0 {
		path, ok = nav.Prev()
	} else {
		path, ok = nav.Next()
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
		if a.diffSource == "branch" {
			a.selectBranchFile(path)
			return
		}
		if a.diffSource == "commit" {
			a.selectCommitFile(context.Background(), a.commitOpen, path)
			return
		}
		a.selectDiff(path)
	} else {
		a.openFile(path)
	}

	if a.fileNav != nil {
		a.fileNav.Select(path)
	}

	a.refresh()
}

// jumpToFirstChange selects the first change even when it is the only one.
func (a *App) jumpToFirstChange(path string) bool {
	rows := a.fileNavChangeRows(path)
	if len(rows) == 0 {
		return false
	}
	group := a.activeGroup()
	if a.split && a.mode != "changes" {
		a.scrollGroupTo(group, rows[0]*rowH)
	} else {
		a.scrollTo(rows[0] * rowH)
	}
	scroll := a.scrollY
	if a.split && a.mode != "changes" {
		scroll = a.scrollGroup[group]
	}
	a.changeNav = changeNavigation{path: path, mode: a.mode, group: group, row: rows[0], scroll: scroll, valid: true}
	a.refresh()
	return true
}

// fileNavOpen opens the current changed file.
func (a *App) fileNavOpen() {
	nav := a.currentFileNav()
	if nav == nil {
		return
	}

	path, ok := nav.Current()
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
