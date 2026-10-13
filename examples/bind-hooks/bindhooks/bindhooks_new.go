package bindhooks

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded template and registers its binding hooks.
// The name starts locked, so the first edit to it is vetoed.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Bind hooks",
		HTML:   bindhooksHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	app.view.Name = "locked"
	app.view.Email = "you@example.com"
	page.Handle(ownframe.Handlers{
		BeforeEdit: app.onBeforeEdit,
		Change:     app.onChange,
	})
	page.SetData(&app.view)

	return app, nil
}
