package clipboard

import "context"

// snap is one saved pair of field values.
type snap struct {
	Left  string
	Right string
}

// snapshot reads both fields. The library owns their values, not View.
func (a *App) snapshot() snap {
	return snap{
		Left:  a.page.Form().FormValue("left"),
		Right: a.page.Form().FormValue("right"),
	}
}

// push saves the current fields for undo. A snapshot equal to the last one
// is dropped so no-op edits do not stack. Any redo history is discarded.
func (a *App) push() {
	next := a.snapshot()
	if n := len(a.undo); n == 0 || a.undo[n-1] != next {
		a.undo = append(a.undo, next)
		if len(a.undo) > undoLimit {
			a.undo = a.undo[len(a.undo)-undoLimit:]
		}
	}

	a.redo = nil
}

// restore writes a snapshot back into the page fields.
func (a *App) restore(s snap) {
	a.page.Form().SetFormValue("left", s.Left)
	a.page.Form().SetFormValue("right", s.Right)
}

func (a *App) onUndo(context.Context) error {
	if len(a.undo) == 0 {
		return nil
	}

	a.redo = append(a.redo, a.snapshot())
	a.restore(a.undo[len(a.undo)-1])
	a.undo = a.undo[:len(a.undo)-1]

	return nil
}

func (a *App) onRedo(context.Context) error {
	if len(a.redo) == 0 {
		return nil
	}

	a.undo = append(a.undo, a.snapshot())
	a.restore(a.redo[len(a.redo)-1])
	a.redo = a.redo[:len(a.redo)-1]

	return nil
}
