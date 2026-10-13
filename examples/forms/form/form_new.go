package form

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded form template and registers its click handler.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Forms",
		HTML:   formHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	page.Handle(ownframe.Handlers{Click: app.onClick})
	page.SetData(app.view)

	return app, nil
}
