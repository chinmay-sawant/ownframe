package dino

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the game page and registers its key handlers and frame tick.
// The game starts on a ready screen until a jump key is pressed.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Dino Run",
		HTML:      buildHTML(),
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
		MaxWidth:  MaxWidth,
		MaxHeight: MaxHeight,
	})
	if err != nil {
		return nil, err
	}

	seed := uint64(time.Now().UnixNano())
	app := &App{
		page:  page,
		game:  newGame(),
		rng:   rand.New(rand.NewPCG(seed, seed^pcgStream)), //nolint:gosec // reason: obstacle shuffle is visual gameplay randomness, not security-sensitive.
		now:   time.Now,
		bound: ^uint64(0),
		view:  view{scale: 1},
	}

	app.clouds = [2]point{{x: cloudAX, y: cloudAY}, {x: cloudBX, y: cloudBY}}

	for i := range app.pebbles {
		app.pebbles[i] = pebbleStart + float64(i)*pebbleStep
	}

	page.Handle(ownframe.Handlers{KeyDown: app.onKeyDown, KeyUp: app.onKeyUp})
	page.SetTick(app.Tick)

	return app, nil
}

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}
