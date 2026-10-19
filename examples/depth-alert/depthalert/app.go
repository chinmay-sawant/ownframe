package depthalert

import (
	"context"
	"sync"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/depth-alert/alert"
)

// App presents the latest Android depth summary in an OwnFrame page.
type App struct {
	page     *ownframe.Page
	mu       sync.Mutex
	analyzer alert.Analyzer
	display  displayFilter
	ops      summaryOps
	view     view
	rendered view
	scanning bool
}

func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{Title: "Depth alert", HTML: pageHTML,
		Width: 360, Height: 280, MinWidth: 320, MinHeight: 260, LockView: true})
	if err != nil {
		return nil, err
	}
	a := &App{page: p}
	a.refreshLocked()
	p.SetData(a.view)
	if err := p.Redraw(context.Background()); err != nil {
		return nil, err
	}
	if err := a.ops.bind(p); err != nil {
		return nil, err
	}
	a.rendered = a.view
	p.SetTick(a.Tick)
	return a, nil
}

func (a *App) Page() *ownframe.Page { return a.page }

// SetScanning clears old measurements at each start and stop boundary.
func (a *App) SetScanning(on bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.scanning = on
	a.display.reset()
	a.analyzer.Reset()
	a.refreshLocked()
}

// ObserveDepth classifies one center-view summary and returns its wire state.
func (a *App) ObserveDepth(nearMM, coverage int32, tracking bool) int32 {
	a.mu.Lock()
	defer a.mu.Unlock()
	state := a.analyzer.Observe(alert.Observation{
		NearestMM: nearMM, Coverage: coverage, Tracking: tracking,
	})
	if a.display.observe(nearMM, coverage, tracking, state) {
		a.refreshLocked()
	}
	return int32(state)
}

func (a *App) Unavailable() int32 {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.analyzer.MarkUnavailable()
	a.display.unavailable()
	a.refreshLocked()
	return int32(a.analyzer.State())
}

func (a *App) refreshLocked() {
	a.view = a.display.view(a.scanning)
}
