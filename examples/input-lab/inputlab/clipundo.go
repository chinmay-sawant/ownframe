package inputlab

import "context"

// undoClip restores the clipboard pair undone last.
func (a *App) undoClip() error {
	if len(a.cUndo) == 0 {
		return nil
	}
	a.cRedo = append(a.cRedo, a.snapshotClip())
	a.restoreClip(a.cUndo[len(a.cUndo)-1])
	a.cUndo = a.cUndo[:len(a.cUndo)-1]
	return nil
}

// redoClip restores the clipboard pair undone last.
func (a *App) redoClip() error {
	if len(a.cRedo) == 0 {
		return nil
	}
	a.cUndo = append(a.cUndo, a.snapshotClip())
	a.restoreClip(a.cRedo[len(a.cRedo)-1])
	a.cRedo = a.cRedo[:len(a.cRedo)-1]
	return nil
}

// selectAllNote refocuses the note before selecting it.
func (a *App) selectAllNote(ctx context.Context) error {
	if !a.focusNote(ctx) {
		a.view.Status = "select-all needs a focused field"
		a.page.SetData(&a.view)
		return nil
	}
	return a.page.SelectAll(ctx)
}

// selectAllClip selects the last clipboard field.
func (a *App) selectAllClip(ctx context.Context) error {
	if !a.focusLast(ctx) {
		a.setClip("nothing to select")
		return nil
	}
	return a.page.SelectAll(ctx)
}
