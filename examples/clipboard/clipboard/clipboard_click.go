package clipboard

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onClick runs the button under the click. A click on a field records it as
// the last field. A button click blurs the form before this runs, so the
// buttons act on the last field the user clicked.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	if box.ID == "left" || box.ID == "right" {
		a.last = box.ID
	}

	switch box.ID {
	case "copy":
		return a.copyButton(ctx)
	case "cut":
		return a.cutButton(ctx)
	case "paste":
		return a.pasteClipboard(ctx)
	case "selectall":
		return a.selectAllButton(ctx)
	case "pastego":
		return a.pasteButton(ctx)
	case "undo":
		return a.Undo(ctx)
	case "redo":
		return a.Redo(ctx)
	}

	return nil
}

func (a *App) copyButton(ctx context.Context) error {
	a.focusLast(ctx)

	text, ok, err := a.CopyText(ctx)
	if err != nil {
		return err
	}

	if !ok || text == "" {
		a.setStatus("nothing selected")

		return nil
	}

	ownframe.Write(text)
	a.setStatus("copied: " + text)

	return nil
}

func (a *App) cutButton(ctx context.Context) error {
	a.focusLast(ctx)

	text, ok, err := a.CutText(ctx)
	if err != nil {
		return err
	}

	if !ok || text == "" {
		a.setStatus("nothing selected")

		return nil
	}

	ownframe.Write(text)
	a.setStatus("cut: " + text)

	return nil
}

func (a *App) selectAllButton(ctx context.Context) error {
	if !a.focusLast(ctx) {
		a.setStatus("nothing to select")

		return nil
	}

	return a.SelectAll(ctx)
}
