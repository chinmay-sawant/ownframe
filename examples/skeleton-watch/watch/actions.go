package watch

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onClick handles pointer clicks on the pushers, winding crown, and lume button.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	now := a.now()
	switch box.ID {
	case "pusher-start", "pusher-top", "btn-chrono-toggle":
		a.chrono.Toggle(now)
		a.view.ChronoRunning = a.chrono.Running()
	case "pusher-reset", "pusher-bottom", "btn-chrono-reset":
		a.chrono.Reset(now)
	case "crown", "btn-wind":
		a.reserve.Wind(WindStepHours)
		a.view.PowerHours = a.reserve.Hours()
		a.view.PowerPercent = a.reserve.Percent()
	case "btn-lume", "lume-toggle":
		a.ToggleLume()
		return a.page.Redraw(ctx)
	}
	return nil
}

// onKeyDown routes Space, R, W, and L shortcuts to watch movement controls.
func (a *App) onKeyDown(ctx context.Context, key string) error {
	now := a.now()
	switch key {
	case "space":
		a.chrono.Toggle(now)
		a.view.ChronoRunning = a.chrono.Running()
	case "r":
		a.chrono.Reset(now)
	case "w":
		a.reserve.Wind(WindStepHours)
		a.view.PowerHours = a.reserve.Hours()
		a.view.PowerPercent = a.reserve.Percent()
	case "l":
		a.ToggleLume()
		return a.page.Redraw(ctx)
	}
	return nil
}
