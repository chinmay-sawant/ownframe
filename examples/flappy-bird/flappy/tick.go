package flappy

import (
	"context"
	"time"
)

// Tick advances the game, drifts the scene, and paints the frame. It runs
// once per window frame before the window draws.
func (a *App) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	now := a.now()
	dt := a.elapse(now)

	a.game.step(dt, a.rng)
	a.stepScene(dt)
	a.paint()

	return nil
}

// elapse returns the seconds since the last tick, capped so a slow frame
// cannot teleport the bird.
func (a *App) elapse(now time.Time) float64 {
	if a.last.IsZero() {
		a.last = now

		return 0
	}

	dt := now.Sub(a.last).Seconds()
	a.last = now

	return min(max(dt, 0), maxDT)
}
