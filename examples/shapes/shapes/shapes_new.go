package shapes

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded shapes template.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Shapes",
		HTML:   shapesHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	app.view.Status = "gallery"
	page.SetData(&app.view)

	return app, nil
}
