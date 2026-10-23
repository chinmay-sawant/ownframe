package inputlab

import "context"

// onUndo routes Ctrl+Z to one section and restores its stack.
// page.Undo only re-invokes this handler, so calling it recurses
// until the stack overflows.
func (a *App) onUndo(context.Context) error {
	switch a.undoGroup() {
	case "si":
		return a.undoLogin()
	case "c":
		return a.undoClip()
	default:
		return a.undoNote()
	}
}

// onRedo routes Ctrl+Y the same way. page.Redo would re-invoke it.
func (a *App) onRedo(context.Context) error {
	switch a.undoGroup() {
	case "si":
		return a.redoLogin()
	case "c":
		return a.redoClip()
	default:
		return a.redoNote()
	}
}

// undoGroup is the focused section, or the last edited one when the
// click on a button blurred the form first.
func (a *App) undoGroup() string {
	if g := focusedGroup(a.page.Edit().FocusedField()); g != "" {
		return g
	}
	return a.lastEdit
}

func focusedGroup(id string) string {
	if id == "e-note" {
		return "e"
	}
	if len(id) > 2 && id[:2] == "c-" {
		return "c"
	}
	if len(id) > 3 && id[:3] == "si-" {
		return "si"
	}
	return ""
}

// pushNote saves e-note with dedup, like the old editing demo.
func (a *App) pushNote() {
	a.lastEdit = "e"
	next := a.page.Form().FormValue("e-note")
	if n := len(a.eUndo); n == 0 || a.eUndo[n-1] != next {
		a.eUndo = append(a.eUndo, next)
		if len(a.eUndo) > undoLimit {
			a.eUndo = a.eUndo[len(a.eUndo)-undoLimit:]
		}
	}
	a.eRedo = nil
}

// undoNote restores the note undone last.
func (a *App) undoNote() error {
	if len(a.eUndo) == 0 {
		a.refresh()
		return nil
	}
	a.eRedo = append(a.eRedo, a.page.Form().FormValue("e-note"))
	a.page.Form().SetFormValue("e-note", a.eUndo[len(a.eUndo)-1])
	a.eUndo = a.eUndo[:len(a.eUndo)-1]
	a.refresh()
	return nil
}

// redoNote restores the note undone last.
func (a *App) redoNote() error {
	if len(a.eRedo) == 0 {
		a.refresh()
		return nil
	}
	a.eUndo = append(a.eUndo, a.page.Form().FormValue("e-note"))
	a.page.Form().SetFormValue("e-note", a.eRedo[len(a.eRedo)-1])
	a.eRedo = a.eRedo[:len(a.eRedo)-1]
	a.refresh()
	return nil
}
