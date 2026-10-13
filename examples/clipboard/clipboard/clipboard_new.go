package clipboard

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the embedded clipboard template and registers its handlers.
// The handlers cover both the buttons and the Ctrl+C/X/V/A/Z/Y chords.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Clipboard",
		HTML:   clipboardHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page, last: "left"}
	page.Handle(ownframe.Handlers{
		Click:      app.onClick,
		BeforeEdit: app.onBeforeEdit,
		Copy:       app.onCopy,
		Cut:        app.onCut,
		Paste:      app.onPaste,
		SelectAll:  app.onSelectAll,
		Undo:       app.onUndo,
		Redo:       app.onRedo,
	})
	page.SetData(app.view)

	return app, nil
}

// onBeforeEdit snapshots the fields so Undo can restore the last edit.
func (a *App) onBeforeEdit(context.Context, ownframe.Box) error {
	a.push()

	return nil
}
