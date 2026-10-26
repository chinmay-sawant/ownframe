package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// onClick routes every control. State changes end in one refresh; the page
// redraws after the handler returns.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	action := box.Action

	switch {
	case action == "":
		if box.ID == "filter" || box.ID == "palette" {
			return nil
		}

		return a.page.Form().Focus(ctx, "")
	case action == "toast-close":
		a.clearToast()
	case action == "mode-explorer":
		a.setMode(ctx, "explorer")
	case action == "mode-changes":
		a.setMode(ctx, "changes")
	case action == "tab-close-working":
		a.closeWorkingTree(ctx)
	case action == "palette-open":
		return a.openPalette(ctx)
	case action == "palette-close":
		return a.closePalette(ctx)
	case action == "diff-refresh":
		a.reloadChanges(ctx, true)
	case action == "diff-note":
		a.showToast("info", "Sample changes", a.changes.Note)
		a.refresh()
	case action == "diff-page-prev":
		if a.changePage > 0 {
			a.changePage--
			a.selectChangePage()
		}
	case action == "diff-page-next":
		if (a.changePage+1)*diffPageRows < len(a.changes.Files) {
			a.changePage++
			a.selectChangePage()
		}
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
	case action == "preview-toggle":
		a.togglePreview()
	case action == "wrap-toggle":
		a.toggleWrap()
	case action == "details-close":
		a.details = false
		a.refresh()
	case action == "collapse-all":
		a.collapseAll()
	case action == "filter-clear":
		a.filter = ""
		a.page.Form().SetFormValue("filter", "")
		a.refresh()
	case action == "focus-filter":
		return a.focusFilter(ctx)
	case action == "copy-path":
		a.copyActive(false)
	case action == "copy-file":
		a.copyActive(true)
	case action == "devtools":
		a.showToast("info", "Developer tools", devMessage(!a.devOn))
		a.SetDevTools(!a.devOn)
	case action == "history-working":
		a.showWorkingDiff()
	case action == "history-prev":
		if a.commitPage > 0 {
			a.commitPage--
			a.refresh()
		}
	case action == "history-next":
		size := a.commitPageSize()
		if (a.commitPage+1)*size < len(a.commits.Commits) {
			a.commitPage++
			a.refresh()
		}
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
		a.toggleFolder(strings.TrimPrefix(action, "dir:"))
	case strings.HasPrefix(action, "tab-close:"):
		if index, ok := actionIndex(action, "tab-close:"); ok {
			a.closeTab(index)
		}
	case strings.HasPrefix(action, "tab:"):
		if index, ok := actionIndex(action, "tab:"); ok {
			a.selectTab(index)
		}
	case strings.HasPrefix(action, "diff-file:"):
		a.selectDiff(strings.TrimPrefix(action, "diff-file:"))
	case strings.HasPrefix(action, "outline:"):
		if line, ok := actionIndex(action, "outline:"); ok {
			a.outlineLine(line)
		}
	}

	return a.page.Form().Focus(ctx, "")
}

// onChange keeps the filter and the palette query in app state after the
// binder writes the fields.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	switch box.ID {
	case "filter":
		next := a.view.Filter
		if next == a.filter {
			return nil
		}

		a.filter = next
		a.refresh()
	case "palette":
		next := a.view.PaletteQuery
		if next == a.paletteQuery {
			return nil
		}

		a.paletteQuery = next
		a.paletteIndex = 0
		a.refresh()
	}

	return nil
}

// onHover shows the commit message card while a history row is hovered.
func (a *App) onHover(_ context.Context, _, next ownframe.Box) (bool, error) {
	if a.historyDrag {
		return false, nil
	}

	short := ""

	if strings.HasPrefix(next.ID, "commit-") &&
		!strings.HasPrefix(next.ID, "commit-file-") &&
		!strings.HasPrefix(next.ID, "commit-open-") {
		short = strings.TrimPrefix(next.ID, "commit-")
	}

	if short == a.hoverCommit {
		return false, nil
	}

	a.hoverCommit = short
	a.refresh()

	return false, nil
}

// onDragStart claims the history divider.
func (a *App) onDragStart(ctx context.Context, box ownframe.Box) (bool, error) {
	if box.ID != "history-divider" || !a.historyVisible() {
		return false, nil
	}

	a.historyDrag = true
	a.historyStart = a.historyRatio
	a.historyDy = 0
	a.hoverCommit = ""

	if err := a.page.Form().Focus(ctx, "history-divider"); err != nil {
		return false, err
	}

	return true, nil
}

// onDragMove resizes the two panes from the drag delta.
func (a *App) onDragMove(ctx context.Context, _, dy float64) error {
	if !a.historyDrag {
		return nil
	}

	a.historyDy += dy

	body := float64(a.bodyHeight() - histSplitH)
	if body <= 0 {
		return nil
	}

	ratio := a.historyStart + a.historyDy/body
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
func (a *App) onDragEnd(context.Context) error {
	a.historyDrag = false
	a.historyDy = 0

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

	if a.ctrl && key == "p" {
		return a.openPalette(ctx)
	}

	if a.ctrl && a.shift && key == "v" {
		a.togglePreview()

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

	if a.page.Form().FocusID() == "filter" {
		switch key {
		case "arrowup", "arrowdown", "pageup", "pagedown":
		default:
			return nil
		}
	}

	if a.page.Form().FocusID() == "history-divider" && a.historyVisible() {
		ratio := a.historyRatio

		switch key {
		case "arrowup":
			ratio -= historyStep
		case "arrowdown":
			ratio += historyStep
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
		a.scrollTo(a.scrollY + scrollStep)
	case "arrowup":
		a.scrollTo(a.scrollY - scrollStep)
	case "pagedown":
		a.scrollTo(a.scrollY + pageScroll)
	case "pageup":
		a.scrollTo(a.scrollY - pageScroll)
	case "home":
		a.scrollTo(0)
	case "end":
		a.scrollTo(a.contentHeight())
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
		if a.details {
			a.details = false
			a.refresh()
		} else if a.toast.Show {
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

// onSubmit opens the selected palette row. Enter and NumpadEnter arrive here.
func (a *App) onSubmit(ctx context.Context) error {
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
	path := a.view.Active
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
