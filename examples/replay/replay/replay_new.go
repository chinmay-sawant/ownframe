package replay

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded pages and registers both routes.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Replay",
		HTML:   replayHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	page.Route("replay", replayHTML)
	page.Route("fallback", fallbackHTML)
	page.SetData(&app.view)

	return app, nil
}
