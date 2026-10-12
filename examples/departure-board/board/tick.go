package board

import (
	"context"
	"time"
)

// rotateEvery is how long one timetable slice stays on the board.
const rotateEvery = 12 * time.Second

// isNight reports whether t is inside the board's night window.
func isNight(t time.Time) bool {
	h := t.Hour()
	return h >= 20 || h < 6
}

// Tick runs once per frame: it opts into the frame-dirty path, retargets the
// feed when the rotation is due, applies the night sheet, binds the retained
// operations after a Redraw, moves the clock and the refresh bar, and steps
// the flaps. It never lays the page out. UseFrameDirty leads every tick
// because the window's FrameDirty call clears the opt-in, and the marks name
// what this frame changed so the window repaints the damage list instead of
// the whole frame.
func (a *App) Tick(ctx context.Context) error {
	a.page.UseFrameDirty()

	if err := ctx.Err(); err != nil {
		return err
	}

	if !a.paused && time.Since(a.lastRotate) >= rotateEvery {
		if err := a.Rotate(ctx); err != nil {
			return err
		}
	}

	if night := isNight(time.Now()); night != a.night {
		if err := a.SetNight(night); err != nil {
			return err
		}

		if err := a.page.Redraw(ctx); err != nil {
			return err
		}
	}

	// A Redraw marked its own dirty region, so rebind names nothing.
	if gen := a.page.Generation(); a.flip.gen != gen || a.flip.clock == nil {
		d := a.page.Display()
		boxes := a.page.Boxes()
		a.flip.rebind(d, boxes, a.view.Rows)
		a.flip.gen = gen
		a.flip.bindBar(d, boxes, a.livery.Accent)
	}

	a.setClockText(formatClock(time.Now()))
	a.moveBar()
	a.invalidateCells(a.flip.step())

	return nil
}
