package flappy

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the game page and registers its input handlers and frame tick.
// The game starts on a ready screen until a flap starts the run.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Flappy Bird",
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
		rng:   rand.New(rand.NewPCG(seed, seed^pcgStream)), //nolint:gosec // reason: deterministic game visuals, not security.
		now:   time.Now,
		bound: ^uint64(0),
	}

	for i := range app.clouds {
		app.clouds[i] = cloudBase + float64(i)*cloudSpacing
	}

	for i := range app.stripes {
		app.stripes[i] = float64(i) * stripeSpacing
	}

	page.Handle(ownframe.Handlers{KeyDown: app.onKeyDown, Click: app.onClick})
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
