package ui

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the template and starts the worker. The window is opened by
// the caller with Run.
func New(opts Options) (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Live log viewer",
		HTML:      buildHTML(),
		Width:     uiWinW,
		Height:    uiWinH,
		MinWidth:  uiMinW,
		MinHeight: uiMinH,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	a := &App{
		page: page, feed: opts.Feed, dark: opts.Dark, perf: opts.Perf,
		exportDir: opts.ExportDir, pollEvery: opts.Poll, ctx: ctx, cancel: cancel,
		reqs: make(chan req, uiReqCap), outs: make(chan out, uiOutCap), started: time.Now(),
	}
	if a.pollEvery <= 0 {
		a.pollEvery = PollEvery
	}

	a.follow = followState{Follow: true}
	a.sourcesDue = time.Now()
	a.tailDue = time.Now().Add(a.pollEvery)
	if a.feed != nil {
		a.pollOn.Store(true)
		a.wg.Add(1)

		go a.run()
		a.loadSettings()
		a.refreshSources()
		a.loadPage(intentNewest)
	}

	page.Handle(ownframe.Handlers{Click: a.onClick, Change: a.onChange, KeyDown: a.onKey})
	page.Scroll().SetWindowing(true)
	page.SetScrollWindow(a.Pin)
	page.SetTick(a.Tick)
	page.SetData(&a.view)
	if err := a.applyTheme(); err != nil {
		cancel()

		return nil, err
	}

	return a, nil
}

// Page returns the ownframe page Run displays.
func (a *App) Page() *ownframe.Page { return a.page }

// Close stops the worker and waits up to ShutdownBudget.
func (a *App) Close() error {
	a.page.SetTick(nil)
	a.cancel()
	done := make(chan struct{})

	go func() { a.wg.Wait(); close(done) }()

	select {
	case <-done:
		return nil
	case <-time.After(ShutdownBudget):
		return ErrShutdown
	}
}
