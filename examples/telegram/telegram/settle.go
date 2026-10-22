package telegram

import (
	"context"
	"time"
)

// settleDelay is how long a phone scroll must hold still before the one
// settling redraw runs.
const settleDelay = 120 * time.Millisecond

// settleScroll redraws once when a phone scroll stops, so the pinned bars
// bake at their final document positions and hit testing matches again.
// While the finger moves, the window blits and the replay draws the pinned
// z-layer at the viewport. The debounce gathers a burst of swipes into one
// redraw after the movement really stops.
func (a *App) settleScroll(ctx context.Context) error {
	if !a.view.Phone {
		a.lastScrollY = -1

		return nil
	}

	_, sy := a.page.Scroll().ScrollOffset()
	now := time.Now()

	if a.lastScrollY == -1 {
		a.lastScrollY = sy

		return nil
	}

	if sy != a.lastScrollY {
		a.lastScrollY, a.lastMove, a.scrollDirty = sy, now, true

		return nil
	}

	if !a.scrollDirty || now.Sub(a.lastMove) < settleDelay {
		return nil
	}

	a.scrollDirty = false

	return a.page.Redraw(ctx)
}
