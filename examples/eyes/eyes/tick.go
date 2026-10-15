package eyes

import "context"

// Tick advances the eyes and paints the frame. It runs once per frame.
func (a *App) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := a.now()
	a.applyMotion(now)
	dt := 1.0 / 60
	if !a.last.IsZero() {
		dt = now.Sub(a.last).Seconds()
		if dt < 0 {
			dt = 0
		}
		if dt > 0.1 {
			dt = 0.1
		}
	}
	a.last = now
	a.step(dt, now)
	a.paint()
	return nil
}
