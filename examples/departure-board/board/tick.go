package board

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// rotateEvery is how long one timetable slice stays on the board.
const rotateEvery = 12 * time.Second

// allCells flattens every row's cells in screen order.
func allCells(v View) []Cell {
	var out []Cell
	for _, row := range v.Rows {
		out = append(out, row.Cells...)
	}
	return out
}

// boxByID returns the first hit-test box with id.
func boxByID(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}
	return ownframe.Box{}, false
}

// isNight reports whether t is inside the board's night window.
func isNight(t time.Time) bool {
	h := t.Hour()
	return h >= 20 || h < 6
}

// Tick runs once per frame: it rotates the feed, applies the night sheet,
// binds the retained operations after a Redraw, moves the clock and the
// refresh bar, and steps the flaps. It never lays the page out.
func (a *App) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !a.paused && time.Since(a.lastRotate) >= rotateEvery {
		if err := a.Advance(ctx); err != nil {
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
	d := a.page.Display()
	boxes := a.page.Boxes()
	gen := a.page.Generation()
	if a.flip.gen != gen || a.flip.clock == nil {
		a.flip.rebind(d, boxes, allCells(a.view))
		a.flip.gen = gen
		a.flip.bindBar(d, boxes, a.livery.Accent)
	}
	if text := formatClock(time.Now()); text != a.flip.clockText {
		a.flip.clockText = text
		a.view.Clock = text
		if a.flip.clock != nil {
			a.flip.clock.Text = text
		}
	}
	if !a.paused && a.flip.bar != nil {
		fraction := float64(time.Since(a.lastRotate)) / float64(rotateEvery)
		fraction = min(max(fraction, 0), 1)
		a.flip.bar.W = a.flip.barFullW * fraction
	}
	a.flip.step()
	return nil
}
