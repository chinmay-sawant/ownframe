package board

import "time"

// setClockText moves the header clock to text and names its box when the text
// changed. Before the first Redraw there is no bound clock op, so there is
// nothing to name.
func (a *App) setClockText(text string) {
	if text == a.flip.clockText {
		return
	}

	a.flip.clockText = text
	a.view.Clock = text

	if a.flip.clock == nil {
		return
	}

	a.flip.clock.Text = text
	a.page.Probe().Invalidate("clock")
}

// barEvery is the refresh fill's cadence: about 10 Hz. A bar this size reads
// smooth at that rate, and the ticks in between report no rect from the bar.
// The cadence is wall-clock driven, so it holds at any rate the window paces
// the page at, not just the 60 ticks per second of the fast loop.
const barEvery = 100 * time.Millisecond

// moveBar advances the refresh fill on the barEvery cadence and names its box
// only when the whole-pixel painted width changed. A paused board holds the
// bar still.
func (a *App) moveBar() {
	if a.paused || a.flip.bar == nil {
		return
	}

	now := time.Now()
	if now.Sub(a.flip.barAt) < barEvery {
		return
	}

	a.flip.barAt = now

	fraction := float64(now.Sub(a.lastRotate)) / float64(rotateEvery)
	fraction = min(max(fraction, 0), 1)

	w := a.flip.barFullW * fraction
	if int(w) == int(a.flip.bar.W) {
		return
	}

	a.flip.bar.W = w
	a.page.Probe().Invalidate("refresh-fill")
}

// invalidateCells marks every tile the flap step changed this frame. Each
// flap holds its box from rebind, so the mark skips the id lookup and its box
// scan.
func (a *App) invalidateCells(ids []string) {
	for _, id := range ids {
		if fl := a.flip.cells[id]; fl != nil {
			a.page.Probe().MarkRect(fl.box)
		}
	}
}
