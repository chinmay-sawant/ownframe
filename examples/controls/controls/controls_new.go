package controls

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded controls template and registers its handlers.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Controls",
		HTML:   controlsHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	page.Handle(ownframe.Handlers{Click: app.onClick, Change: app.onChange})
	page.SetData(app.view)

	return app, nil
}
