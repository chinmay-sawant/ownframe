package pet

import (
	"context"
)

// Tick is called once per frame by the window loop.
func (a *App) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	dt := 1.0 / 60.0
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

	oldE := int(a.state.Energy + 0.5)
	oldH := int(a.state.Hunger + 0.5)
	oldHp := int(a.state.Happiness + 0.5)
	oldMood := a.state.Mood

	a.state.Decay(dt, now)

	newE := int(a.state.Energy + 0.5)
	newH := int(a.state.Hunger + 0.5)
	newHp := int(a.state.Happiness + 0.5)

	if oldE != newE || oldH != newH || oldHp != newHp || oldMood != a.state.Mood ||
		now.After(a.state.SpeechUntil) {
		a.refreshView(now)
	}

	a.paint(now)
	return nil
}
