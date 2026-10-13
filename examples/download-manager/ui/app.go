package ui

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// The window sizes. The minimum keeps the table usable; a smaller frame is
// refused by the host, and a larger page scrolls.
const (
	DefaultWidth  = 1180
	DefaultHeight = 780
	MinWidth      = 780
	MinHeight     = 520
)

// App is the download-manager screen.
type App struct {
	page    *ownframe.Page
	backend Backend
	view    View
	pager   *Pager
	detail  *Row

	bound        uint64
	bindings     bindings
	geom         bool
	needActive   bool
	started      bool
	historyDirty bool
	footWait     bool
	sumWait      bool
	detailWait   bool
	summaryGen   uint64
	lastSummary  time.Time
	lastHistory  time.Time
}

// New builds the page from the backend mode and theme, registers the
// handlers and the tick, and starts the backend workers.
func New(ctx context.Context, backend Backend) (*App, error) {
	dark, err := backend.Dark()
	if err != nil {
		dark = false
	}

	info := backend.Info()
	app := &App{
		backend: backend,
		pager:   NewPager(HistoryPage),
	}
	app.view = View{Dark: dark, Dummy: info.Dummy, DataDir: info.DataDir}

	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Download Manager",
		HTML:      buildHTML(),
		Theme:     themeSource(dark),
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}

	app.page = page
	page.Handle(ownframe.Handlers{
		Click:  app.onClick,
		Change: app.onChange,
		Submit: app.onSubmit,
	})
	page.SetData(app.view)
	page.SetTick(app.Tick)
	backend.Start(ctx)

	return app, nil
}
