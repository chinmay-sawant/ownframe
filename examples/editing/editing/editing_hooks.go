package editing

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onClick dispatches the plain buttons to the matching page action.
// The redraw button bumps the stamp; the page redraws after the handler.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	switch box.ID {
	case "selectall":
		return a.selectAllButton(ctx)
	case "undo":
		return a.page.Edit().Undo(ctx)
	case "redo":
		return a.page.Edit().Redo(ctx)
	case "redraw":
		a.view.Stamp++
	}

	return a.refresh()
}

// onBeforeEdit saves the note for undo before the edit runs.
func (a *App) onBeforeEdit(context.Context, ownframe.Box) error {
	a.push()

	return nil
}

// onChange refreshes the status line after the note changed.
func (a *App) onChange(context.Context, ownframe.Box) error {
	return a.refresh()
}

// onSelectAll runs when a shortcut selects with no focused field.
// It refocuses the note so the chord still selects it.
func (a *App) onSelectAll(ctx context.Context) error {
	if !a.focusNote(ctx) {
		a.view.Status = "select-all needs a focused field"
		a.page.SetData(a.view)

		return nil
	}

	return a.page.Edit().SelectAll(ctx)
}

// onUndo restores the previous note.
func (a *App) onUndo(context.Context) error {
	if len(a.undo) == 0 {
		return a.refresh()
	}

	a.redo = append(a.redo, a.page.Form().FormValue("note"))
	a.page.Form().SetFormValue("note", a.undo[len(a.undo)-1])
	a.undo = a.undo[:len(a.undo)-1]

	return a.refresh()
}

// onRedo restores the note undone last.
func (a *App) onRedo(context.Context) error {
	if len(a.redo) == 0 {
		return a.refresh()
	}

	a.undo = append(a.undo, a.page.Form().FormValue("note"))
	a.page.Form().SetFormValue("note", a.redo[len(a.redo)-1])
	a.redo = a.redo[:len(a.redo)-1]

	return a.refresh()
}
