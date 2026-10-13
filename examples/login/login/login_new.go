package login

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded login template and registers its handlers.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Sign in",
		HTML:      loginHTML,
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	page.Handle(ownframe.Handlers{
		Click:      app.onClick,
		Submit:     app.onSubmit,
		BeforeEdit: app.onBeforeEdit,
		Change:     app.onChange,
		Undo:       app.onUndo,
		Redo:       app.onRedo,
	})
	page.SetData(app.view)

	return app, nil
}
