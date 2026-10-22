package login

import "context"

func (a *App) onUndo(context.Context) error {
	if len(a.undo) == 0 {
		return nil
	}

	a.redo = append(a.redo, a.snapshot())
	a.restore(a.undo[len(a.undo)-1])
	a.undo = a.undo[:len(a.undo)-1]
	a.refresh()

	return nil
}

func (a *App) onRedo(context.Context) error {
	if len(a.redo) == 0 {
		return nil
	}

	a.undo = append(a.undo, a.snapshot())
	a.restore(a.redo[len(a.redo)-1])
	a.redo = a.redo[:len(a.redo)-1]
	a.refresh()

	return nil
}

func (a *App) onSubmit(_ context.Context) error {
	if !a.ready() {
		return nil
	}

	a.signIn()
	a.page.SetData(&a.view)

	return nil
}
