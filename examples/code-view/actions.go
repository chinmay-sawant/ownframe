package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// onClick routes every control. State changes end in one refresh; the page

// onClick routes every control. State changes end in one refresh; the page
// redraws after the handler returns.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	if a.dismissOpenMenu(box.ID, box.Action) {
		return nil
	}

	if a.confirmOpen {
		switch box.Action {
		case "confirm-ok", "confirm-cancel":
		default:
			if box.ID != "confirm-backdrop" {
				return nil
			}
		}
	}

	action := box.Action

	switch {
	case action == "":
		if box.ID == "filter" || box.ID == "palette" || box.ID == "find" || box.ID == "branch-filter" ||
			box.ID == "workspace-search" || box.ID == "editor" || box.ID == "edit-surface" || box.ID == "project-path" ||
			box.ID == "drawer-path" {
			return nil
		}

		if a.inlineCard.Show {
			a.closeInlineCard()
		}

		if a.fileMenu && !strings.HasPrefix(box.ID, "file-") &&
			!strings.HasPrefix(box.ID, "project-") && !strings.HasPrefix(box.ID, "recent-") &&
			!strings.HasPrefix(box.ID, "browse-") && !strings.HasPrefix(box.ID, "quick-") {
			a.fileMenu = false
			a.projectOpen = false
			a.recentOpen = false
			a.refresh()
		}

		return a.page.Form().Focus(ctx, "")
	case action == "inline-card-close":
		a.closeInlineCard()
	case action == "inline-card-prev":
		a.inlineCardStep(-1)
	case action == "inline-card-next":
		a.inlineCardStep(1)
	case strings.HasPrefix(action, "inline-card:"):
		if line, ok := actionIndex(action, "inline-card:"); ok {
			a.showInlineCard(line)
		}
	case action == "toast-close":
		a.clearToast()
	case action == "split-toggle":
		a.toggleSplit()
	case strings.HasPrefix(action, "focus-group:"):
		if index, ok := actionIndex(action, "focus-group:"); ok {
			a.setFocusGroup(index)
		}
	case action == "edit-begin":
		a.inlineCard = inlineCard{}
		if !a.editing {
			a.toggleEdit()
		}

		return a.page.Redraw(ctx)
	case strings.HasPrefix(action, "edit-begin-line:"):
		if line, ok := actionIndex(action, "edit-begin-line:"); ok {
			// Press already placed the caret with its column; a Click that
			// follows must not reset it to the line start.
			if a.editing && a.edit != nil {
				if doc := a.groupDoc(0); doc != nil && a.edit.Path == doc.Path {
					if curLine, _ := offsetLineCol(a.edit.Text(), a.editCaret); curLine == line {
						return a.page.Redraw(ctx)
					}
				}
			}
			a.beginEditAtLine(line)
		}

		return a.page.Redraw(ctx)
	case action == "edit-toggle":
		a.toggleEdit()
	case action == "edit-save":
		a.saveEdit()
	case action == "diff-edit":
		a.editDiffFile(-1)

		return a.page.Redraw(ctx)
	case strings.HasPrefix(action, "diff-edit-line:"):
		if line, ok := actionIndex(action, "diff-edit-line:"); ok && line > 0 {
			// Press already opened the file with its column; a Click that
			// follows must not reset it to the line start.
			if a.editing && a.edit != nil && a.edit.Path == a.diffPath {
				if curLine, _ := offsetLineCol(a.edit.Text(), a.editCaret); curLine == line-1 {
					return a.page.Redraw(ctx)
				}
			}
			a.editDiffFile(line - 1)
		}

		return a.page.Redraw(ctx)
	case action == "search-open":
		return a.openWorkspaceSearch(ctx)
	case action == "file-nav-prev":
		a.fileNavStep(-1)
	case action == "file-nav-next":
		a.fileNavStep(1)
	case action == "file-nav-change-prev":
		a.fileNavChangeStep(-1)
	case action == "file-nav-change-next":
		a.fileNavChangeStep(1)
	case action == "file-nav-open":
		a.fileNavOpen()
	case action == "nav-back":
		a.goNavBack()
	case action == "nav-forward":
		a.goNavForward()
	case action == "search-close":
		return a.closeWorkspaceSearch(ctx)
	case action == "search-case":
		a.toggleSearchCase(ctx)

		return a.page.Form().Focus(ctx, "workspace-search")
	case action == "search-prev":
		a.searchStep(-1)
	case action == "search-next":
		a.searchStep(1)
	case strings.HasPrefix(action, "search-hit:"):
		if index, ok := actionIndex(action, "search-hit:"); ok {
			a.openSearchHit(index)
		}
	case action == "mode-explorer":
		a.setMode(ctx, "explorer")
	case action == "mode-changes":
		a.setMode(ctx, "changes")
	case action == "tab-close-working":
		a.closeWorkingTree(ctx)
	case action == "settings-open":
		a.settingsOpen = true
		a.paletteOpen = false
		a.searchOpen = false
		a.branchOpen = false
		a.refresh()
		return a.page.Form().Focus(ctx, "")
	case action == "settings-close":
		a.settingsOpen = false
		a.refresh()
		return a.page.Form().Focus(ctx, "act-settings")
	case action == "file-menu-open":
		a.fileMenu = !a.fileMenu
		if !a.fileMenu {
			a.projectOpen = false
			a.recentOpen = false
			a.browseOpen = false
		}
		a.paletteOpen = false
		a.searchOpen = false
		a.branchOpen = false
		a.settingsOpen = false
		a.refresh()

		if a.fileMenu {
			return a.page.Form().Focus(ctx, "drawer-path")
		}

		return a.page.Form().Focus(ctx, "")
	case action == "file-menu-close":
		a.fileMenu = false
		a.projectOpen = false
		a.recentOpen = false
		a.browseOpen = false
		a.fileAlert = ""
		a.refresh()
		return a.page.Form().Focus(ctx, "")
	case action == "file-reset":
		a.resetLayout()
		return a.page.Form().Focus(ctx, "")
	case action == "file-reload-window":
		return a.reloadWindow(ctx)
	case action == "file-new-window":
		a.fileMenu = false
		a.projectOpen = false
		a.browseOpen = false
		a.refresh()
		return a.openNewWindow()
	case action == "project-open":
		a.projectOpen = !a.projectOpen
		a.refresh()
		return a.page.Form().Focus(ctx, "drawer-path")
	case action == "project-close":
		a.projectOpen = false
		a.refresh()
		return a.page.Form().Focus(ctx, "")
	case action == "recent-toggle":
		a.recentOpen = !a.recentOpen
		a.refresh()
		return nil
	case action == "project-submit":
		root := strings.TrimSpace(a.projectQuery)
		if root == "" {
			root = strings.TrimSpace(a.view.ProjectQuery)
		}
		if root == "" {
			return nil
		}
		if blockedProjectRoot(root) {
			a.fileAlert = "Drive roots can't be opened as projects. Pick a folder inside it."
			a.refresh()
			return nil
		}
		a.browseOpen = false
		return a.switchProject(ctx, root)
	case strings.HasPrefix(action, "project-open-path:"):
		root := strings.TrimSpace(strings.TrimPrefix(action, "project-open-path:"))
		if root == "" {
			return nil
		}
		if blockedProjectRoot(root) {
			a.fileAlert = "Drive roots can't be opened as projects. Pick a folder inside it."
			a.refresh()
			return nil
		}
		a.browseOpen = false
		return a.switchProject(ctx, root)
	case action == "browse-open":
		a.fileMenu = false
		a.projectOpen = false
		a.recentOpen = false
		a.startBrowse()
		a.refresh()
		return a.page.Form().Focus(ctx, "project-path")
	case action == "browse-close", action == "browse-backdrop":
		a.browseOpen = false
		a.refresh()
		return nil
	case action == "browse-up":
		next, ok := browseUpTarget(a.browseRoot)
		if !ok {
			return nil
		}
		a.browseRoot = next
		a.browseList.reset()
		a.fileAlert = ""
		a.refresh()
		return nil
	case action == "browse-wsl":
		a.browseRoot = a.browseStartRoot()
		if a.browseRoot == "" {
			return nil
		}
		a.browseList.reset()
		a.fileAlert = ""
		a.refresh()
		return nil
	case action == "browse-windows":
		for c := 'c'; c <= 'z'; c++ {
			candidate := "/mnt/" + string(c)
			if isDir(candidate) {
				a.browseRoot = candidate
				a.browseList.reset()
				a.fileAlert = ""
				a.refresh()
				return nil
			}
		}
		if isDir("/mnt") {
			a.browseRoot = "/mnt"
			a.browseList.reset()
			a.fileAlert = ""
			a.refresh()
			return nil
		}
		a.fileAlert = "No Windows drives found under /mnt."
		a.refresh()
		return nil
	case strings.HasPrefix(action, "browse-enter:"):
		path := strings.TrimSpace(strings.TrimPrefix(action, "browse-enter:"))
		if path == "" {
			return nil
		}
		if !isDir(path) {
			a.fileAlert = "Folder not found: " + path
			a.refresh()
			return nil
		}
		a.browseRoot = path
		a.browseList.reset()
		a.fileAlert = ""
		a.refresh()
		return nil
	case action == "browse-select":
		if a.browseRoot == "" {
			return nil
		}
		if blockedProjectRoot(a.browseRoot) {
			a.fileAlert = "Drive roots can't be opened as projects. Pick a folder inside it."
			a.refresh()
			return nil
		}
		a.browseOpen = false
		return a.switchProject(ctx, a.browseRoot)
	case action == "file-alert-close":
		a.fileAlert = ""
		a.refresh()
		return nil
	case strings.HasPrefix(action, "theme:"):
		name := validTheme(strings.TrimPrefix(action, "theme:"))
		a.theme = name
		if err := a.page.SetTheme(a.activeThemeCSS()); err != nil {
			return err
		}
		a.refresh()
	case action == "palette-open":
		return a.openPalette(ctx)
	case action == "win-min":
		a.page.Window().Minimize()

		return nil
	case action == "win-max":
		a.page.Window().ToggleMaximize()
		a.refresh()

		return nil
	case action == "zoom-reset":
		return a.resetZoom(ctx)
	case action == "win-close":
		a.page.Window().Close()

		return nil
	case action == "palette-close":
		return a.closePalette(ctx)
	case action == "diff-refresh":
		a.reloadChanges(ctx, true)
	case action == "stage-all":
		if err := a.stageAll(ctx); err != nil {
			return err
		}
	case action == "unstage-all":
		if err := a.unstageAll(ctx); err != nil {
			return err
		}
	case action == "discard-all":
		a.openDiscardConfirm("discard-all", a.workingAllPaths())
	case action == "stage-selected":
		if err := a.stageSelected(ctx); err != nil {
			return err
		}
	case action == "unstage-selected":
		if err := a.unstageSelected(ctx); err != nil {
			return err
		}
	case action == "discard-selected":
		a.openDiscardConfirm("discard-selected", a.selectedList())
	case action == "select-all-files":
		a.selectAllFiles()
		a.refresh()
	case action == "select-clear":
		if a.diffPath != "" {
			a.setSingleSelect(a.diffPath)
		} else {
			a.ensureSelected()
			for key := range a.selected {
				delete(a.selected, key)
			}
		}
		a.refresh()
	case action == "confirm-ok":
		if err := a.confirmDiscard(ctx); err != nil {
			return err
		}
	case action == "confirm-cancel":
		a.closeConfirm()
	case strings.HasPrefix(action, "stage-file:"):
		if err := a.stageFile(ctx, strings.TrimPrefix(action, "stage-file:")); err != nil {
			return err
		}
	case strings.HasPrefix(action, "unstage-file:"):
		if err := a.unstageFile(ctx, strings.TrimPrefix(action, "unstage-file:")); err != nil {
			return err
		}
	case strings.HasPrefix(action, "discard-file:"):
		a.openDiscardConfirm("discard-file", []string{strings.TrimPrefix(action, "discard-file:")})
	case strings.HasPrefix(action, "copy-commit-message:"):
		a.copyCommitMessage(strings.TrimPrefix(action, "copy-commit-message:"))
	case strings.HasPrefix(action, "copy-commit-id:"):
		a.copyCommitID(strings.TrimPrefix(action, "copy-commit-id:"))
	case action == "diff-note":
		a.showToast("info", "Sample changes", a.changes.Note)
		a.refresh()
	case action == "diff-page-prev":
		a.scrollChanges(-1)
	case action == "diff-page-next":
		a.scrollChanges(1)
	case action == "diff-split":
		a.setSplit(true)
	case action == "diff-unified":
		a.setSplit(false)
	case action == "hunk-prev":
		a.jumpHunk(-1)
	case action == "hunk-next":
		a.jumpHunk(1)
	case action == "details-toggle":
		a.toggleDetails()
	case action == "scroll-left":
		a.scrollHorizontal(-a.zx(hscrollStep))
	case action == "scroll-right":
		a.scrollHorizontal(a.zx(hscrollStep))
	case action == "tabs-prev":
		a.scrollTabs(0, -1)
	case action == "tabs-next":
		a.scrollTabs(0, 1)
	case action == "tabs2-prev":
		a.scrollTabs(1, -1)
	case action == "tabs2-next":
		a.scrollTabs(1, 1)
	case action == "preview-toggle":
		a.togglePreview()
	case action == "wrap-toggle":
		a.toggleWrap()
	case action == "find-close":
		return a.closeFind(ctx)
	case action == "find-prev":
		a.findStep(-1)

		return a.page.Form().Focus(ctx, "find")
	case action == "find-next":
		a.findStep(1)

		return a.page.Form().Focus(ctx, "find")
	case action == "find-case":
		a.toggleFindCase()

		return a.page.Form().Focus(ctx, "find")
	case action == "details-close":
		a.details = false
		a.refresh()
	case action == "collapse-all":
		a.collapseAll()
	case action == "filter-clear":
		a.filter = ""
		a.treeOffset = 0
		a.page.Form().SetFormValue("filter", "")
		a.refresh()
	case action == "focus-filter":
		return a.focusFilter(ctx)
	case action == "copy-path":
		a.copyActive(false)
	case action == "copy-file":
		a.copyActive(true)
	case strings.HasPrefix(action, "menu-item:"):
		return a.runMenuItem(ctx, strings.TrimPrefix(action, "menu-item:"))
	case action == "devtools":
		a.showToast("info", "Developer tools", devMessage(!a.devOn))
		a.SetDevTools(!a.devOn)
	case action == "branch-picker":
		a.branchOpen = !a.branchOpen
		a.branchQuery = ""

		if !a.branchOpen {
			a.refresh()

			break
		}

		a.refresh()
		if err := a.page.Redraw(ctx); err != nil {
			return err
		}

		return a.page.Form().Focus(ctx, "branch-filter")
	case action == "branch-close":
		a.branchOpen = false
		a.branchQuery = ""
		a.refresh()
	case action == "branch-head":
		a.pickBranch(ctx, "")
	case strings.HasPrefix(action, "branch-pick:"):
		if index, ok := actionIndex(action, "branch-pick:"); ok && index < len(a.view.History.Branches) {
			a.pickBranch(ctx, a.view.History.Branches[index].Ref)
		}
	case action == "tree-up":
		a.scrollTree(-(a.treeVisibleRows() - 1))
	case action == "tree-down":
		a.scrollTree(a.treeVisibleRows() - 1)
	case action == "history-working":
		a.showWorkingDiff()
	case action == "history-branch-changes":
		a.showBranchChanges(ctx)
	case action == "history-prev":
		a.scrollHistory(ctx, -1)
	case action == "history-next":
		a.scrollHistory(ctx, 1)
	case strings.HasPrefix(action, "commit-file:"):
		rest := strings.TrimPrefix(action, "commit-file:")
		if parts := strings.SplitN(rest, ":", actionPair); len(parts) == actionPair {
			a.selectCommitFile(ctx, parts[0], parts[1])
		}
	case strings.HasPrefix(action, "commit-open:"):
		a.openCommitDiff(ctx, strings.TrimPrefix(action, "commit-open:"))
	case strings.HasPrefix(action, "commit:"):
		a.toggleCommit(ctx, strings.TrimPrefix(action, "commit:"))
	case strings.HasPrefix(action, "open:"):
		a.openFile(strings.TrimPrefix(action, "open:"))
	case strings.HasPrefix(action, "dir:"):
		a.onToggleDir(strings.TrimPrefix(action, "dir:"))
	case strings.HasPrefix(action, "tab-close:"):
		if index, ok := actionIndex(action, "tab-close:"); ok {
			a.closeTab(index)
		}
	case strings.HasPrefix(action, "tab2-close:"):
		if index, ok := actionIndex(action, "tab2-close:"); ok {
			a.closeTabIn(1, index)
		}
	case strings.HasPrefix(action, "tab2:"):
		if index, ok := actionIndex(action, "tab2:"); ok {
			a.selectTabIn(1, index)
		}
	case strings.HasPrefix(action, "tab:"):
		if index, ok := actionIndex(action, "tab:"); ok {
			a.selectTab(index)
		}
	case strings.HasPrefix(action, "diff-file:"):
		path := strings.TrimPrefix(action, "diff-file:")
		switch a.diffSource {
		case "commit":
			a.selectCommitFile(ctx, a.commitOpen, path)
		case "branch":
			a.selectBranchFile(path)
		default:
			if a.ctrl {
				a.toggleDiff(path)
			} else {
				a.selectDiff(path)
			}
		}
	case strings.HasPrefix(action, "diff-expand-up:"):
		if idx, ok := actionIndex(action, "diff-expand-up:"); ok {
			a.diffExpandUp(idx)
		}
	case strings.HasPrefix(action, "diff-expand-down:"):
		if idx, ok := actionIndex(action, "diff-expand-down:"); ok {
			a.diffExpandDown(idx)
		}
	case strings.HasPrefix(action, "diff-expand-all:"):
		if idx, ok := actionIndex(action, "diff-expand-all:"); ok {
			a.diffExpandAll(idx)
		}
	case strings.HasPrefix(action, "outline:"):
		if line, ok := actionIndex(action, "outline:"); ok {
			a.outlineLine(line)
		}
	}

	return a.page.Form().Focus(ctx, "")
}

// onChange keeps the filter and the palette query in app state after the
// binder writes the fields.
func (a *App) onChange(ctx context.Context, box ownframe.Box) error {
	switch box.ID {
	case "filter":
		next := a.view.Filter
		if next == a.filter {
			return nil
		}

		a.filter = next
		a.treeOffset = 0
		a.refresh()
	case "find":
		next := a.view.FindQuery
		if next == a.findQuery {
			return nil
		}

		a.findQuery = next
		a.reindexFind()
		a.refresh()
	case "branch-filter":
		next := a.view.BranchQuery
		if next == a.branchQuery {
			return nil
		}

		a.branchQuery = next
		a.refresh()
	case "palette":
		next := a.view.PaletteQuery
		if next == a.paletteQuery {
			return nil
		}

		a.paletteQuery = next
		a.paletteIndex = 0
		a.refresh()
	case "editor":
		a.applyEditText(a.view.EditText)
	case "workspace-search":
		next := a.view.SearchQuery
		if next == a.searchQuery {
			return nil
		}

		a.searchQuery = next
		a.searchIndex = 0
		a.runWorkspaceSearch(ctx)
	case "project-path", "drawer-path":
		next := a.view.ProjectQuery
		if next == a.projectQuery {
			return nil
		}

		a.projectQuery = next
		a.refresh()
	}

	return nil
}

func findBoxID(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}

	return ownframe.Box{}, false
}

// onWheel scrolls the explorer tree, the changed-file list, the history list,
// or the unwrapped editor under the pointer. A vertical wheel over the editor
// returns false so the window can pan a long file. Ctrl+wheel does not zoom
// this page.
func (a *App) onWheel(ctx context.Context, x, y, dx, dy float64) (bool, error) {
	if a.confirmOpen {
		return true, nil
	}

	// Ctrl+wheel zooms the layout like the keyboard chords. The window
	// leaves the wheel to the page because wheel zoom is off.
	if a.ctrl && dy != 0 {
		if dy > 0 {
			return true, a.zoomIn(ctx)
		}

		return true, a.zoomOut(ctx)
	}

	if a.searchOpen {
		sx, sy := a.page.Scroll().ScrollOffset()
		x -= float64(sx)
		y -= float64(sy)
		box, ok := findBoxID(a.page.Boxes(), "search-results")
		if ok && dy != 0 && x >= box.X && x <= box.X+box.W && y >= box.Y && y <= box.Y+box.H {
			rows := a.searchScroll.WheelRows(-dy * treeWheelRows)
			if rows != 0 {
				a.scrollSearchTo(a.searchScroll.Offset + rows)
			}
			return true, a.page.Redraw(ctx)
		}
		return true, nil
	}
	if a.settingsOpen {
		return true, nil
	}
	if a.paletteOpen || a.branchOpen {
		return false, nil
	}

	scrollX, scrollY := a.page.Scroll().ScrollOffset()
	x -= float64(scrollX)
	y -= float64(scrollY)

	if handled, err := a.browseWheelAt(ctx, x, y, dy); handled || err != nil {
		return handled, err
	}

	if !a.wrap && ((dx != 0) || (a.shift && dy != 0)) && a.isOverEditor(x, y) {
		delta := -dx
		if dx == 0 && a.shift && dy != 0 {
			delta = -dy
		}

		step := int(delta * float64(hscrollWheelStep))
		if step == 0 && delta != 0 {
			if delta > 0 {
				step = 1
			} else {
				step = -1
			}
		}

		a.scrollHorizontal(step)

		return true, a.page.Redraw(ctx)
	}

	if dy == 0 {
		return false, nil
	}

	if a.mode == "changes" && a.view.Diff.Scroll {
		box, ok := findBoxID(a.page.Boxes(), "changes-list")
		if ok && x >= box.X && x <= box.X+box.W && y >= box.Y && y <= box.Y+box.H {
			rows := a.changesScroll.WheelRows(-dy * treeWheelRows)
			if rows != 0 {
				a.scrollChanges(rows)
			}

			return true, a.page.Redraw(ctx)
		}
	}

	if a.mode == "changes" && a.view.History.Visible {
		box, ok := findBoxID(a.page.Boxes(), "history-pane")
		if ok && x >= box.X && x <= box.X+box.W && y >= box.Y && y <= box.Y+box.H {
			rows := a.historyScroll.WheelRows(-dy * treeWheelRows)
			if rows != 0 {
				a.scrollHistory(ctx, rows)
			}

			return true, a.page.Redraw(ctx)
		}
	}

	if a.mode == "explorer" {
		if a.split {
			if a.isOverPane(1, x, y) {
				a.scrollEditorGroupByWheel(1, dy)

				return true, a.page.Redraw(ctx)
			}

			if a.isOverPane(0, x, y) {
				a.scrollEditorGroupByWheel(0, dy)

				return true, a.page.Redraw(ctx)
			}
		}

		if !a.view.TreeScroll {
			return false, nil
		}

		tree, ok := findBoxID(a.page.Boxes(), "tree-pane")
		if !ok {
			return false, nil
		}

		if x < tree.X || x > tree.X+tree.W || y < tree.Y || y > tree.Y+tree.H {
			a.treeWheelAccum = 0

			return false, nil
		}

		a.treeWheelAccum += -dy * treeWheelRows
		rows := int(a.treeWheelAccum)
		a.treeWheelAccum -= float64(rows)

		if rows != 0 {
			a.scrollTree(rows)
			a.keepTreeFocus(ctx)
		}

		return true, a.page.Redraw(ctx)
	}

	return false, nil
}

// keepTreeFocus moves focus to the first visible row when a wheel scroll
// carried the focused row out of the window.
func (a *App) keepTreeFocus(ctx context.Context) {
	focus := a.page.Form().FocusID()
	if !strings.HasPrefix(focus, "tree-") {
		return
	}

	path := strings.TrimPrefix(focus, "tree-")

	for _, row := range a.view.Tree {
		if row.Path == path {
			return
		}
	}

	if len(a.view.Tree) > 0 {
		_ = a.page.Form().Focus(ctx, "tree-"+a.view.Tree[0].Path)
	}
}

// scrollTree moves the explorer window by rows.
func (a *App) scrollTree(rows int) {
	a.treeOffset += rows
	a.clampTreeOffset(a.view.TreeTotal, a.treeVisibleRows())
	a.refresh()
}

// onHover shows the commit message card while a history row is hovered.
func (a *App) onHover(_ context.Context, _, next ownframe.Box) (bool, error) {
	if a.historyDrag {
		return false, nil
	}

	short := ""
	path := ""

	switch {
	case strings.HasPrefix(next.ID, "tree-"):
		path = strings.TrimPrefix(next.ID, "tree-")
	case strings.HasPrefix(next.ID, "tab-") && !strings.HasPrefix(next.ID, "tab-close-") && next.ID != "tab-details":
		if index, ok := actionIndex(next.ID, "tab-"); ok && index < len(a.tabs) {
			path = a.tabs[index]
		}
	case strings.HasPrefix(next.ID, "tab2-") && !strings.HasPrefix(next.ID, "tab2-close-"):
		if index, ok := actionIndex(next.ID, "tab2-"); ok && index < len(a.tabs2) {
			path = a.tabs2[index]
		}
	}

	if strings.HasPrefix(path, "close") || strings.HasPrefix(path, "working") || strings.HasPrefix(path, "details") {
		path = ""
	}

	if path != "" {
		if doc := a.docs[path]; doc != nil {
			if path != a.hoverPath {
				a.hoverPath = path
				a.refresh()
			}

			return false, nil
		}

		for _, file := range a.ws.Files {
			if file.Path == path {
				if path != a.hoverPath {
					a.hoverPath = path
					a.refresh()
				}

				return false, nil
			}
		}
	}

	if strings.HasPrefix(next.ID, "commit-") &&
		!strings.HasPrefix(next.ID, "commit-file-") &&
		!strings.HasPrefix(next.ID, "commit-open-") {
		short = strings.TrimPrefix(next.ID, "commit-")
	}

	if short == a.hoverCommit && path == "" && a.hoverPath == "" {
		return false, nil
	}

	if short == a.hoverCommit && a.hoverPath != "" {
		a.hoverPath = ""
		a.refresh()
	}

	a.hoverCommit = short
	a.hoverPath = ""
	a.refresh()

	return false, nil
}

// treeKeyDown moves focus through the visible explorer rows.
func (a *App) treeKeyDown(ctx context.Context, key string) (bool, error) {
	focus := a.page.Form().FocusID()
	if focus == "tree-up" || focus == "tree-down" {
		switch key {
		case "arrowdown", "pagedown":
			a.scrollTree(a.treeVisibleRows() - 1)
		case "arrowup", "pageup":
			a.scrollTree(-(a.treeVisibleRows() - 1))
		default:
			return false, nil
		}

		a.refresh()

		return true, a.page.Redraw(ctx)
	}

	if !strings.HasPrefix(focus, "tree-") || focus == "tree-scroll" {
		return false, nil
	}

	path := strings.TrimPrefix(focus, "tree-")
	index := -1

	for i, row := range a.view.Tree {
		if row.Path == path {
			index = i

			break
		}
	}

	if index < 0 {
		return false, nil
	}

	switch key {
	case "arrowdown":
		if index+1 >= len(a.view.Tree) {
			a.scrollTree(1)
			if err := a.page.Redraw(ctx); err != nil {
				return true, err
			}

			index = 0
		} else {
			index++
		}
	case "arrowup":
		if index == 0 {
			a.scrollTree(-1)
			if err := a.page.Redraw(ctx); err != nil {
				return true, err
			}

			index = 0
		} else {
			index--
		}
	case "pagedown":
		a.scrollTree(a.treeVisibleRows() - 1)
		if err := a.page.Redraw(ctx); err != nil {
			return true, err
		}

		index = 0
	case "pageup":
		a.scrollTree(-(a.treeVisibleRows() - 1))
		if err := a.page.Redraw(ctx); err != nil {
			return true, err
		}

		index = len(a.view.Tree) - 1
	case "home":
		a.treeOffset = 0
		a.refresh()

		if err := a.page.Redraw(ctx); err != nil {
			return true, err
		}

		index = 0
	case "end":
		a.treeOffset = max(0, a.view.TreeTotal-a.treeVisibleRows())
		a.refresh()

		if err := a.page.Redraw(ctx); err != nil {
			return true, err
		}

		index = len(a.view.Tree) - 1
	case "arrowright":
		if a.folders[path] {
			return true, nil
		}

		if a.view.Tree[index].Dir {
			a.onToggleDir(path)
			if err := a.page.Redraw(ctx); err != nil {
				return true, err
			}
		}

		return true, a.page.Form().Focus(ctx, focus)
	case "arrowleft":
		if a.folders[path] {
			a.onToggleDir(path)
			if err := a.page.Redraw(ctx); err != nil {
				return true, err
			}

			return true, a.page.Form().Focus(ctx, focus)
		}

		parent := dirName(path)
		if parent == "." || parent == "" {
			return true, nil
		}

		visible := false

		for _, row := range a.view.Tree {
			if row.Path == parent {
				visible = true

				break
			}
		}

		if !visible {
			index := 0
			if treeRowIndex(a.ws.Tree, a.folders, parent, &index) {
				a.treeOffset = max(0, index-treeParentPad)
				a.refresh()

				if err := a.page.Redraw(ctx); err != nil {
					return true, err
				}
			}
		}

		return true, a.page.Form().Focus(ctx, "tree-"+parent)
	default:
		return false, nil
	}

	if index >= 0 && index < len(a.view.Tree) {
		return true, a.page.Form().Focus(ctx, "tree-"+a.view.Tree[index].Path)
	}

	return true, nil
}

// onDragStart claims the history divider or the explorer thumb.
func (a *App) onDragStart(ctx context.Context, box ownframe.Box) (bool, error) {
	if box.ID == "editor-scroll-thumb" || box.ID == "editor-scroll-thumb-2" {
		group := 0
		if box.ID == "editor-scroll-thumb-2" {
			group = 1
		}
		return a.startEditorScrollDrag(group), nil
	}
	if box.ID == "search-thumb" && a.searchOpen && a.view.Search.Scroll {
		track := searchWindowRows * a.zx(searchRowHeight)
		thumb := a.view.Search.ScrollHeight
		if track <= thumb {
			return false, nil
		}
		a.searchDrag = true
		a.searchDragAt = a.searchScroll.Offset
		a.searchDragScale = float64(len(a.searchHits)-searchWindowRows) / float64(track-thumb)
		return true, nil
	}
	if a.onDragClaim(box) {
		return true, nil
	}

	if box.ID == "history-thumb" && a.mode == "changes" && a.view.History.Scroll {
		visible := a.historyVisibleRows()
		track := a.historyTrackHeight()
		thumb := a.view.History.ScrollHeight
		total := len(a.commits.Commits)
		if total <= visible || track <= thumb {
			return false, nil
		}

		a.historyListDrag = true
		a.historyListAt = a.historyScroll.Offset
		a.historyListScale = float64(total-visible) / float64(track-thumb)

		return true, nil
	}

	if box.ID == "changes-thumb" && a.mode == "changes" && a.view.Diff.Scroll {
		visible := a.changeVisibleRows()
		track := a.changesTrackHeight()
		thumb := a.view.Diff.ScrollHeight
		total := len(a.sidebarChangeFiles())
		if total <= visible || track <= thumb {
			return false, nil
		}

		a.changesDrag = true
		a.changesDragAt = a.changesScroll.Offset
		a.changesScale = float64(total-visible) / float64(track-thumb)

		return true, nil
	}

	if box.ID == "tree-thumb" && a.mode == "explorer" && a.view.TreeScroll {
		visible := a.treeVisibleRows()
		track := visible * a.zx(treeRowH)
		thumb := a.view.TreeThumbH
		total := a.view.TreeTotal

		if total <= visible || track <= thumb {
			return false, nil
		}

		a.treeDrag = true
		a.treeDragAt = a.treeOffset
		a.treeScale = float64(total-visible) / float64(track-thumb)

		return true, nil
	}

	if box.ID != "history-divider" || !a.historyVisible() {
		return false, nil
	}

	a.historyDrag = true
	a.historyStart = float64(a.historyHeight()) / float64(a.drawerStackHeight())
	a.historyManual = true
	a.historyDy = 0
	a.hoverCommit = ""

	if err := a.page.Form().Focus(ctx, "history-divider"); err != nil {
		return false, err
	}

	return true, nil
}

// onDragMove resizes the panes or moves the explorer thumb.
func (a *App) onDragMove(ctx context.Context, dx, dy float64) error {
	if a.editorDrag.active {
		a.moveEditorScrollDrag(dy)
		return a.page.Redraw(ctx)
	}
	if a.searchDrag {
		a.scrollSearchTo(a.searchDragAt + int(dy*a.searchDragScale+roundHalf))
		return a.page.Redraw(ctx)
	}
	if a.dragSelect {
		a.selX += dx
		a.selY += dy
		a.extendEditSelection(a.selX, a.selY)
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if a.historyListDrag {
		offset := a.historyListAt + int(dy*a.historyListScale+roundHalf)
		total := len(a.commits.Commits)
		visible := a.historyVisibleRows()
		a.historyScroll.SetOffset(offset, total, visible)
		a.maybeLoadMoreHistory(ctx)
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if a.changesDrag {
		offset := a.changesDragAt + int(dy*a.changesScale+roundHalf)
		total := len(a.sidebarChangeFiles())
		visible := a.changeVisibleRows()
		a.changesScroll.SetOffset(offset, total, visible)
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if a.treeDrag {
		offset := a.treeDragAt + int(dy*a.treeScale+roundHalf)
		maxOffset := a.view.TreeTotal - a.treeVisibleRows()
		if maxOffset < 0 {
			maxOffset = 0
		}

		if offset < 0 {
			offset = 0
		}

		if offset > maxOffset {
			offset = maxOffset
		}

		if offset == a.treeOffset {
			return nil
		}

		a.treeOffset = offset
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if !a.historyDrag {
		return nil
	}

	a.historyDy += dy

	body := float64(a.drawerStackHeight())
	if body <= 0 {
		return nil
	}

	ratio := a.historyStart - a.historyDy/body
	if ratio < historyMinRatio {
		ratio = historyMinRatio
	}

	if ratio > historyMaxRatio {
		ratio = historyMaxRatio
	}

	if ratio == a.historyRatio {
		return nil
	}

	a.historyRatio = ratio
	a.refresh()

	return a.page.Redraw(ctx)
}

// onDragEnd releases the history divider.
func (a *App) onDragEnd(ctx context.Context) error {
	a.editorDrag.active = false
	wasSelect := a.dragSelect
	wasSearch := a.searchDrag
	a.dragSelect = false
	a.searchDrag = false
	a.searchDragScale = 0
	a.historyDrag = false
	a.historyDy = 0
	a.treeDrag = false
	a.treeScale = 0
	a.changesDrag = false
	a.changesScale = 0
	a.historyListDrag = false
	a.historyListScale = 0
	if wasSearch && a.searchOpen {
		return a.page.Form().Focus(ctx, "workspace-search")
	}

	if wasSelect {
		a.refresh()

		return a.page.Redraw(ctx)
	}

	return nil
}

// onKeyDown maps the single-key shortcuts and tracks the modifiers that the
// window sends as separate key events.
func (a *App) onKeyDown(ctx context.Context, key string) error {
	switch key {
	case "control", "controlleft", "controlright", "meta", "metaleft", "metaright":
		a.ctrl = true

		return nil
	case "shift", "shiftleft", "shiftright":
		a.shift = true

		return nil
	}

	// Layout zoom owns Ctrl+=, Ctrl+-, and Ctrl+0 everywhere, including
	// while editing: the window leaves these chords to the page because
	// key zoom is off. Copy, cut, paste, and the other Ctrl chords keep
	// their own branches below.
	if a.ctrl {
		switch key {
		case "equal", "kpadd":
			return a.zoomIn(ctx)
		case "minus", "kpsubtract":
			return a.zoomOut(ctx)
		case "0", "kp0":
			return a.resetZoom(ctx)
		}
	}

	// VS Code-style history: Alt+Left goes back, Alt+Right goes forward. The
	// window sends these as single names, so they work while editing and win
	// over the horizontal scroll below.
	switch key {
	case "alt+arrowleft":
		a.goNavBack()

		return a.page.Redraw(ctx)
	case "alt+arrowright":
		a.goNavForward()

		return a.page.Redraw(ctx)
	}

	if a.confirmOpen && key == "escape" {
		a.closeConfirm()

		return a.page.Redraw(ctx)
	}

	if a.confirmOpen && key == "enter" {
		if err := a.confirmDiscard(ctx); err != nil {
			return err
		}

		return a.page.Redraw(ctx)
	}

	if a.confirmOpen {
		return nil
	}

	if a.settingsOpen && key == "escape" {
		a.settingsOpen = false
		a.refresh()
		if err := a.page.Redraw(ctx); err != nil {
			return err
		}
		return a.page.Form().Focus(ctx, "act-settings")
	}

	if a.menuOpen && key == "escape" {
		a.closeMenu()
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if a.browseOpen && key == "escape" {
		if a.fileAlert != "" {
			a.fileAlert = ""
			a.refresh()

			return a.page.Redraw(ctx)
		}
		a.browseOpen = false
		a.refresh()
		if err := a.page.Redraw(ctx); err != nil {
			return err
		}
		return a.page.Form().Focus(ctx, "file-menu")
	}

	if a.fileMenu && key == "escape" {
		if a.fileAlert != "" {
			a.fileAlert = ""
			a.refresh()

			return a.page.Redraw(ctx)
		}
		if a.recentOpen || a.projectOpen {
			a.recentOpen = false
			a.projectOpen = false
			a.refresh()

			return a.page.Redraw(ctx)
		}

		a.fileMenu = false
		a.projectOpen = false
		a.recentOpen = false
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if !a.paletteOpen && (key == "enter" || key == "space") && a.page.Edit().FocusedField() == "" {
		if id := a.page.Form().FocusID(); id != "" {
			if box, ok := findBoxID(a.page.Boxes(), id); ok {
				if err := a.onClick(ctx, box); err != nil {
					return err
				}

				if a.page.Form().FocusID() == "" {
					_ = a.page.Form().Focus(ctx, id)
				}

				return a.page.Redraw(ctx)
			}
		}
	}

	if a.settingsOpen {
		return nil
	}

	if a.editing {
		return a.editKey(ctx, key)
	}

	if a.ctrl && key == "a" && !a.shift {
		if a.mode == "changes" && a.confirmOpen == false {
			switch a.page.Form().FocusID() {
			case "filter", "palette", "find", "branch-filter", "workspace-search", "project-path", "drawer-path", "editor", "edit-surface":
			default:
				a.selectAllFiles()
				a.refresh()

				return a.page.Redraw(ctx)
			}
		}

		return nil
	}

	if a.ctrl && key == "p" {
		return a.openPalette(ctx)
	}

	if a.ctrl && a.shift && key == "f" {
		return a.openWorkspaceSearch(ctx)
	}

	if a.ctrl && key == "f" {
		return a.openFind(ctx)
	}

	if a.ctrl && key == "s" {
		a.saveDirtyActive()

		return a.page.Redraw(ctx)
	}

	if a.ctrl && key == "e" {
		a.toggleEdit()

		return a.page.Redraw(ctx)
	}

	if a.ctrl && key == "backslash" {
		a.toggleSplit()

		return a.page.Redraw(ctx)
	}

	if a.ctrl && a.shift && key == "v" {
		a.togglePreview()

		return a.page.Redraw(ctx)
	}

	if a.searchOpen && key == "escape" {
		return a.closeWorkspaceSearch(ctx)
	}

	if a.page.Form().FocusID() == "workspace-search" {
		switch key {
		case "arrowdown":
			a.searchStep(1)

			return a.page.Redraw(ctx)
		case "arrowup":
			a.searchStep(-1)

			return a.page.Redraw(ctx)
		case "enter":
			a.openSearchHit(a.searchIndex)

			return a.page.Redraw(ctx)
		case "escape":
			return a.closeWorkspaceSearch(ctx)
		default:
			return nil
		}
	}

	if a.branchOpen && key == "escape" {
		a.branchOpen = false
		a.branchQuery = ""
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if a.paletteOpen {
		switch key {
		case "arrowdown":
			a.movePalette(1)

			return a.page.Redraw(ctx)
		case "arrowup":
			a.movePalette(-1)

			return a.page.Redraw(ctx)
		case "escape":
			return a.closePalette(ctx)
		case "enter":
			return a.onSubmit(ctx)
		default:
			return nil
		}
	}

	if a.page.Form().FocusID() == "find" {
		switch key {
		case "escape":
			return a.closeFind(ctx)
		case "enter":
			a.findStep(1)

			return a.page.Redraw(ctx)
		default:
			return nil
		}
	}

	if handled, err := a.treeKeyDown(ctx, key); handled || err != nil {
		return err
	}

	if a.page.Form().FocusID() == "branch-filter" {
		switch key {
		case "escape":
			a.branchOpen = false
			a.branchQuery = ""
			a.refresh()

			return a.page.Redraw(ctx)
		case "enter":
			if len(a.view.History.Branches) > 0 {
				a.pickBranch(ctx, a.view.History.Branches[0].Ref)
			}

			return a.page.Redraw(ctx)
		default:
			return nil
		}
	}

	if a.page.Form().FocusID() == "filter" {
		switch key {
		case "arrowup", "arrowdown", "pageup", "pagedown":
		default:
			return nil
		}
	}

	if a.page.Form().FocusID() == "history-divider" && a.historyVisible() {
		if !a.historyManual {
			a.historyRatio = float64(a.historyHeight()) / float64(a.drawerStackHeight())
		}

		ratio := a.historyRatio

		switch key {
		case "arrowup":
			ratio += historyStep
		case "arrowdown":
			ratio -= historyStep
		case "home":
			ratio = historyDefaultRatio
		default:
			return nil
		}

		if ratio < historyMinRatio {
			ratio = historyMinRatio
		}

		if ratio > historyMaxRatio {
			ratio = historyMaxRatio
		}

		a.historyRatio = ratio
		a.historyManual = true
		a.refresh()

		return a.page.Redraw(ctx)
	}

	switch key {
	case "1":
		a.setMode(ctx, "explorer")
	case "2":
		a.setMode(ctx, "changes")
	case "3":
		a.toggleDetails()
	case "arrowdown":
		a.scrollFocusedGroup(a.zx(scrollStep))
	case "arrowup":
		a.scrollFocusedGroup(-a.zx(scrollStep))
	case "arrowleft":
		if !a.wrap {
			a.scrollHorizontal(-a.zx(hscrollStep))

			return a.page.Redraw(ctx)
		}
	case "arrowright":
		if !a.wrap {
			a.scrollHorizontal(a.zx(hscrollStep))

			return a.page.Redraw(ctx)
		}
	case "pagedown":
		a.scrollFocusedGroup(a.zx(pageScroll))
	case "pageup":
		a.scrollFocusedGroup(-a.zx(pageScroll))
	case "home":
		a.scrollFocusedGroupTo(0)
	case "end":
		a.scrollFocusedGroupToEnd()
	case "f":
		return a.focusFilter(ctx)
	case "w":
		if a.mode == "changes" {
			a.closeWorkingTree(ctx)
		} else {
			a.closeTab(a.active)
		}
	case "z":
		a.toggleWrap()
	case "m":
		a.togglePreview()

		return a.page.Redraw(ctx)
	case "r":
		if a.mode == "changes" {
			a.reloadChanges(ctx, true)
		} else {
			a.reloadActive(ctx)
		}
	case "s":
		if a.mode == "changes" {
			a.setSplit(true)
		}
	case "u":
		if a.mode == "changes" {
			a.setSplit(false)
		}
	case "n":
		if a.mode == "changes" {
			a.jumpHunk(1)
		}
	case "p":
		if a.mode == "changes" {
			a.jumpHunk(-1)
		}
	case "escape":
		switch {
		case a.inlineCard.Show:
			a.closeInlineCard()
		case a.details:
			a.details = false
			a.refresh()
		case a.toast.Show:
			a.clearToast()
		}
	default:
		return nil
	}

	return a.page.Redraw(ctx)
}

// onKeyUp clears the tracked modifiers and closes the palette on the release
// so the input blur does not swallow the close.
func (a *App) onKeyUp(ctx context.Context, key string) error {
	switch key {
	case "control", "controlleft", "controlright", "meta", "metaleft", "metaright":
		a.ctrl = false
	case "shift", "shiftleft", "shiftright":
		a.shift = false
	case "escape":
		if a.paletteOpen {
			return a.closePalette(ctx)
		}
	}

	return nil
}

// onSubmit opens the selected palette row or advances find. Enter and
// NumpadEnter arrive here.
func (a *App) onSubmit(ctx context.Context) error {
	if a.editing {
		return a.insertEditText(ctx, "\n")
	}

	if (a.fileMenu || a.browseOpen) && (a.page.Form().FocusID() == "project-path" ||
		a.page.Form().FocusID() == "drawer-path") {
		root := strings.TrimSpace(a.view.ProjectQuery)
		if root == "" {
			root = strings.TrimSpace(a.projectQuery)
		}
		if root == "" {
			return nil
		}
		if blockedProjectRoot(root) {
			a.fileAlert = "Drive roots can't be opened as projects. Pick a folder inside it."
			a.refresh()
			return nil
		}
		return a.switchProject(ctx, root)
	}

	if a.findOpen && a.page.Form().FocusID() == "find" {
		delta := 1
		if a.shift {
			delta = -1
		}

		a.findStep(delta)

		return a.page.Redraw(ctx)
	}

	if !a.paletteOpen || a.page.Form().FocusID() != "palette" {
		return nil
	}

	rows := a.paletteRows()
	if len(rows) == 0 {
		return nil
	}

	index := a.paletteIndex
	if index < 0 {
		index = 0
	}

	if index >= len(rows) {
		index = len(rows) - 1
	}

	a.openFile(rows[index].Path)

	if err := a.page.Form().Focus(ctx, ""); err != nil {
		return err
	}

	return a.page.Redraw(ctx)
}

// focusFilter switches to the explorer and focuses the filter input.
func (a *App) focusFilter(ctx context.Context) error {
	if a.mode != "explorer" {
		a.setMode(ctx, "explorer")

		if err := a.page.Redraw(ctx); err != nil {
			return err
		}
	}

	return a.page.Form().Focus(ctx, "filter")
}

// copyActive puts the active path, or the whole file, on the clipboard.
func (a *App) copyActive(wholeFile bool) {
	path := a.activeDocPath()
	if path == "" {
		return
	}

	if !wholeFile {
		ownframe.Write(path)
		a.showToast("success", "Path copied", path)
		a.refresh()

		return
	}

	raw, err := a.ws.readFile(path)
	if err != nil {
		a.showToast("error", "Cannot copy", err.Error())
		a.refresh()

		return
	}

	ownframe.Write(string(raw))
	a.showToast("success", "File copied", path)
	a.refresh()
}

// reloadActive re-reads the active file from the workspace.
func (a *App) reloadActive(_ context.Context) {
	doc := a.activeDoc()
	if doc == nil {
		a.showToast("info", "No file open", "Open a file before reloading it.")
		a.refresh()

		return
	}

	next := a.ws.openDoc(doc.Path)
	a.docs[doc.Path] = next

	if next.Err != "" {
		a.showToast("error", "Reload failed", next.Err)
	} else {
		a.showToast("success", "File reloaded", doc.Path)
	}

	a.refresh()
}

// actionIndex parses the number after an action prefix.
func actionIndex(action, prefix string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimPrefix(action, prefix))
	if err != nil || n < 0 {
		return 0, false
	}

	return n, true
}

// devMessage is the toast copy for the dock toggle.
func devMessage(on bool) string {
	if on {
		return "The dock is open. F12 or Ctrl+Shift+I closes it."
	}

	return "The dock is closed. F12 or Ctrl+Shift+I opens it."
}

// onToggleDir handles expanding or collapsing a directory in the tree.
// When expanding an unloaded directory, its immediate children are loaded from disk.
func (a *App) onToggleDir(path string) {
	a.folders[path] = !a.folders[path]

	if a.folders[path] {
		if node := findTreeNode(a.ws.Tree, path); node != nil && node.Dir && !node.Loaded {
			_ = a.ws.loadNode(node)
		} else {
			_ = a.ws.loadDir(path)
		}
		a.revealTreeFolder(path)
	}

	a.refresh()
}
