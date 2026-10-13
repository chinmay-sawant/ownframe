package drop

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded template and registers the drop handler.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Drag and drop",
		HTML:   dropHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	page.Handle(ownframe.Handlers{Drop: app.onDrop, Change: app.onChange})
	page.SetData(&app.view)

	return app, nil
}
