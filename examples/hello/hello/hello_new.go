package hello

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded hello template.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{
		Title:  "Hello",
		HTML:   helloHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page, view: View{Message: "Hello, World!"}}
	page.SetData(&app.view)

	return app, nil
}
