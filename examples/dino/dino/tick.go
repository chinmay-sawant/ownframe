package dino

import (
	"context"
	"time"
)

// Tick advances the game, measures the frame rate, and paints the scene.
// It runs once per window frame before the window draws.
func (a *App) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	a.syncView()

	now := a.now()
	dt := a.elapse(now)

	a.game.step(dt, a.rng)
	a.tapStep()
	a.duckStep(now)
	a.stepSky(dt)
	a.fps.add(now)
	a.paint()

	return nil
}

// elapse returns the seconds since the last tick, capped so a slow frame
// cannot teleport the game.
func (a *App) elapse(now time.Time) float64 {
	if a.last.IsZero() {
		a.last = now

		return 0
	}

	dt := now.Sub(a.last).Seconds()
	a.last = now

	return min(max(dt, 0), maxFrameDT)
}
