package inputlab

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// onClick dispatches buttons and links. Field clicks record the
// clipboard side; button clicks blur first so they use it.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	if box.ID == "c-left" || box.ID == "c-right" {
		a.lastClip = box.ID
	}
	if box.Action == "send" {
		return a.send()
	}
	if box.Action == "login" || box.ID == "si-login" {
		if a.ready() {
			a.signIn()
		}
		a.page.SetData(&a.view)
		return nil
	}
	switch {
	case box.ID == "i-more":
		a.scrollTo(0, scrollEnd)
		return nil
	case strings.HasPrefix(box.ID, "e-"):
		return a.editClick(ctx, box.ID)
	case strings.HasPrefix(box.ID, "c-"):
		return a.clipClick(ctx, box.ID)
	}
	a.refresh()
	return nil
}

func (a *App) editClick(ctx context.Context, id string) error {
	switch id {
	case "e-selectall":
		return a.selectAllNote(ctx)
	case "e-undo":
		return a.undoNote()
	case "e-redo":
		return a.redoNote()
	case "e-redraw":
		a.view.Stamp++
	}
	a.refresh()
	return nil
}
