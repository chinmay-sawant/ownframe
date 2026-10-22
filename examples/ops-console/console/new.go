package console

import "github.com/chinmay-sawant/ownframe"

// New parses the shell template and registers one Handlers value.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{
		DevTools:  true,
		Title:     "Yard Ledger",
		HTML:      shellHTML,
		Theme:     dayTheme,
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page, tickets: seedTickets()}
	app.view.Screen = screenLogin
	app.view.Site = sitePier
	app.view.Filter = filterAll
	app.syncView()
	page.Handle(ownframe.Handlers{
		Click:  app.onClick,
		Change: app.onChange,
		Submit: app.onSubmit,
	})
	page.SetData(&app.view)

	return app, nil
}
