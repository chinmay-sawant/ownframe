package editing

import "fmt"

// push saves the current note for undo and drops any redo history.
// A snapshot equal to the last one is dropped so no-op edits do not stack.
func (a *App) push() {
	next := a.page.Form().FormValue("note")
	if n := len(a.undo); n == 0 || a.undo[n-1] != next {
		a.undo = append(a.undo, next)
		if len(a.undo) > undoLimit {
			a.undo = a.undo[len(a.undo)-undoLimit:]
		}
	}

	a.redo = nil
}

// refresh prints the note state on the status line and stores the data.
func (a *App) refresh() error {
	a.view.Status = fmt.Sprintf(
		"focus=%s value=%q undo=%d redo=%d",
		a.page.Edit().FocusedField(),
		a.page.Form().FormValue("note"),
		len(a.undo),
		len(a.redo),
	)
	a.page.SetData(a.view)

	return nil
}
