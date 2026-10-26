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
		return nil
	case action == "toast-close":
		a.clearToast()
	case action == "mode-explorer":
		a.setMode(ctx, "explorer")
	case action == "mode-changes":
		a.setMode(ctx, "changes")
	case action == "diff-refresh":
		a.reloadChanges(ctx, true)
	case action == "diff-note":
		a.showToast("info", "Sample changes", a.changes.Note)
		a.refresh()
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
			a.scrollTo(line * rowH)
		}
	}

	return nil
}

// onChange keeps the filter in app state after the binder writes the field.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	if box.ID != "filter" {
		return nil
	}

	next := a.view.Filter
	if next == a.filter {
		return nil
	}

	a.filter = next
	a.refresh()

	return nil
}

// onKeyDown maps the single-key shortcuts. It leaves every key alone while
// the filter input has focus, so typing stays typing.
func (a *App) onKeyDown(ctx context.Context, key string) error {
	if a.page.Form().FocusID() == "filter" {
		return nil
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
		a.closeTab(a.active)
	case "r":
		a.reloadActive(ctx)
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
		} else if a.toast.Show {
			a.toast = toastView{}
		}
	default:
		return nil
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
		return
	}

	a.docs[doc.Path] = a.ws.openDoc(doc.Path)
	a.showToast("success", "File reloaded", doc.Path)
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
