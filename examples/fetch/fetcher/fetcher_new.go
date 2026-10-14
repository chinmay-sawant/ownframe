package fetcher

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded template and registers its click handler.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Fetch",
		HTML:   fetcherHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	page.Handle(ownframe.Handlers{Click: app.onClick})
	page.SetData(&app.view)

	return app, nil
}
