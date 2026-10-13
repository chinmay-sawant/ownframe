package inputlab

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded template and registers every handler.
// Bound fields need a pointer so data-bind writes into View.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Input lab",
		HTML:      inputlabHTML,
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}
	app := &App{page: page, lastClip: "c-left"}
	app.view.LName = "locked"
	app.view.LEmail = "you@example.com"
	app.view.BColor = "Red"
	app.view.Status = "hover, press, focus, or check a control"
	page.Handle(ownframe.Handlers{
		Click:      app.onClick,
		Change:     app.onChange,
		BeforeEdit: app.onBeforeEdit,
		Submit:     app.onSubmit,
		Copy:       app.onCopy,
		Cut:        app.onCut,
		Paste:      app.onPaste,
		SelectAll:  app.onSelectAll,
		Undo:       app.onUndo,
		Redo:       app.onRedo,
	})
	page.SetData(&app.view)
	return app, nil
}
