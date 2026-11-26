package ui

import (
	"context"
	"slices"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// modalFocus tracks keyboard focus across a modal's lifetime: where focus came
// from, so closing returns there, and which modal the Tab scope belongs to.
type modalFocus struct {
	last   string // modal shown at the end of the previous redraw
	opener string // focused control when the first modal opened
	scoped bool
}

// modalContainers names the element that wraps each modal's controls.
func (a *App) modalContainer() string {
	if a.modal == "palette" {
		return "palette"
	}
	return "modal"
}

// modalFirstFocus is the control a modal focuses when it opens: the first field
// for forms, the safe button for confirmations.
func (a *App) modalFirstFocus() string {
	switch a.modal {
	case "conn":
		return "cd-name"
	case "export":
		return "ex-fmt-csv"
	case "import":
		return "im-path"
	case "history":
		return "hist-search"
	case "settings":
		return "set-theme-dark"
	case "sqlpreview":
		return "pv-close"
	case "confirm":
		if a.confirm != nil && a.confirm.Typed != "" {
			return "confirm-typed"
		}
		return "confirm-cancel"
	}
	return ""
}

// noteModalOpener runs before a redraw builds the tree. The first redraw with a
// modal open remembers the focused control, before the modal moves focus.
func (a *App) noteModalOpener() {
	if a.modal != "" && a.mf.last == "" && a.page != nil {
		a.mf.opener = a.page.Form().FocusID()
	}
}

// syncModalFocus runs after a redraw. While a modal is open it limits Tab and
// Shift+Tab to the modal's controls and moves focus in on open; when the last
// modal closes it lifts the limit and returns focus to the opener.
func (a *App) syncModalFocus(ctx context.Context) error {
	form := a.page.Form()
	switch {
	case a.modal != "":
		ids := a.modalControlIDs()
		if len(ids) > 0 {
			form.SetFocusScope(ids...)
			a.mf.scoped = true
		}
		if a.mf.last != a.modal && a.modal != "palette" {
			if first := a.modalFirstFocus(); first != "" && slices.Contains(ids, first) {
				if err := form.Focus(ctx, first); err != nil {
					return err
				}
			}
		}
	case a.mf.last != "":
		if a.mf.scoped {
			form.SetFocusScope()
			a.mf.scoped = false
		}
		opener := a.mf.opener
		a.mf.opener = ""
		if opener != "" && a.hasBox(opener) {
			a.mf.last = ""
			return form.Focus(ctx, opener)
		}
	}
	a.mf.last = a.modal
	return nil
}

func (a *App) hasBox(id string) bool {
	for _, b := range a.page.Boxes() {
		if b.ID == id {
			return true
		}
	}
	return false
}

// modalControlIDs lists the ids of every box inside the open modal, in document
// order. Ids that are not controls are skipped by the page's Tab traversal.
func (a *App) modalControlIDs() []string {
	if a.modal == "palette" {
		return []string{"pal-input"} // the rows are picked by arrow keys, the scrollbar by pointer
	}
	boxes := a.page.Boxes()
	var host ownframe.Box
	want := a.modalContainer()
	for _, b := range boxes {
		if b.ID == want {
			host = b
			break
		}
	}
	if host.ID == "" {
		return nil
	}
	var ids []string
	for _, b := range boxes {
		if b.ID == "" || b.ID == want || strings.HasSuffix(b.ID, "-thumb") || strings.HasSuffix(b.ID, "-track") {
			continue
		}
		if b.X >= host.X && b.Y >= host.Y && b.X+b.W <= host.X+host.W+1 && b.Y+b.H <= host.Y+host.H+1 {
			ids = append(ids, b.ID)
		}
	}
	return ids
}

// modalTextInputs returns the ids of text fields in dialogs; Enter in one of
// them submits the dialog.
func modalTextInputs() []string {
	return []string{"cd-name", "cd-path", "cd-host", "cd-port", "cd-db", "cd-user", "cd-pass", "cd-dsn",
		"ex-delim", "ex-null", "ex-path", "im-path", "im-target", "set-slow", "confirm-typed"}
}

// modalEnter runs the open dialog's primary action. The page calls Submit only
// when no button has focus, so Enter on a focused button still presses that
// button.
func (a *App) modalEnter(ctx context.Context) error {
	focus := a.page.Form().FocusID()
	if focus != "" && !slices.Contains(modalTextInputs(), focus) {
		return nil
	}
	switch a.modal {
	case "conn":
		if focus == "cd-dsn" {
			return a.fillFromDSN(ctx)
		}
		return a.connectDraft(ctx)
	case "export":
		return a.runExport(ctx, true)
	case "import":
		if a.importSt.Data == nil {
			return a.previewImport(ctx)
		}
		return a.runImport(ctx)
	case "settings":
		a.applySlowThreshold()
		return a.closeModal(ctx)
	case "confirm":
		return a.confirmOK(ctx)
	case "sqlpreview", "history":
		return a.closeModal(ctx)
	}
	return nil
}
