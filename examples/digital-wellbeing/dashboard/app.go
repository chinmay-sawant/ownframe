// Package dashboard is Mochi, an Android screen-time companion.
package dashboard

import (
	"context"
	_ "embed"
	"sync"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed screen.html
var screen string

//go:embed assets/mochi.png
var mochi []byte

// App owns the page and queues Android data for the page loop.
type App struct {
	page *ownframe.Page
	view View
	bridge
	selectionMu   sync.RWMutex
	selected      string
	selecting     bool
	weekOffset    int
	currentScreen string
	returnScreen  string
}

// New builds the app and registers the original cat illustration.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{
		Title:  "Digital Wellbeing with Mochi",
		HTML:   screen,
		Width:  windowWidth,
		Height: windowHeight,
	})
	if err != nil {
		return nil, err
	}
	app := &App{page: page}
	app.clearUsage()
	app.refreshView()
	page.SetImage("mochi", mochi)
	page.Handle(ownframe.Handlers{Click: app.click})
	page.SetData(&app.view)
	page.SetTick(app.Tick)
	return app, nil
}

// Page returns the page for a native or desktop host.
func (a *App) Page() *ownframe.Page { return a.page }

// Tick applies queued data from Android on the game loop.
func (a *App) Tick(ctx context.Context) error { return a.apply(ctx) }
