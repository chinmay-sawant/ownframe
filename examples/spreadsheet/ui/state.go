package ui

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// shutdownBudget is the documented time Close waits for the worker to join.
const shutdownBudget = 2 * time.Second

// New builds the screen: page, handlers, tick, and the first draw.
func New(opts Options) (*App, error) {
	if opts.Backend == nil {
		return nil, errNilBackend
	}

	a := &App{
		b:     opts.Backend,
		opts:  opts,
		info:  map[string]Sheet{},
		sel:   map[string]Selection{},
		used:  map[string]Area{},
		tiles: newTileCache(),
	}

	title := opts.Title
	if title == "" {
		title = "Spreadsheet"
	}

	width, height := opts.Width, opts.Height
	if width <= 0 {
		width = 1180
	}

	if height <= 0 {
		height = 760
	}

	p, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     title,
		HTML:      pageHTML,
		Width:     width,
		Height:    height,
		MinWidth:  uiMinW,
		MinHeight: uiMinH,
		Perf:      opts.Perf,
	})
	if err != nil {
		return nil, err
	}

	a.page = p
	a.loadSheets()
	if v, ok := a.b.Pref("theme"); ok && v == "dark" {
		a.dark = true
	}

	a.work = newWorker(a.b)
	p.SetData(&a.view)
	p.Handle(a.handlers())
	p.SetTick(a.tick)
	p.SetWindowing(true)
	p.SetScrollWindow(a.scrollWindow)
	if err := a.Redraw(context.Background()); err != nil {
		return nil, err
	}

	return a, nil
}

// loadSheets reads the sheet list and the current revision.
func (a *App) loadSheets() {
	a.sheets = a.b.Sheets()
	for _, sh := range a.sheets {
		a.info[sh.ID] = sh
	}

	if len(a.sheets) > 0 {
		a.active = a.sheets[0].ID
	}

	a.rev = a.b.Revision()
}

// Close stops the worker within the shutdown budget and closes the
// backend. It is safe to call twice.
func (a *App) Close() error {
	if a.closed {
		return nil
	}

	a.closed = true
	a.work.close(shutdownBudget)

	return a.b.Close()
}
