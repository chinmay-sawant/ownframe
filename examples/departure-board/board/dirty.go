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
	a.page.Invalidate("clock")
}

// barEvery is how many ticks between refresh fill updates: about 10 Hz at the
// window's 60 ticks per second. A bar this size reads smooth at that rate,
// and the ticks in between report no rect from the bar.
const barEvery = 6

// moveBar advances the refresh fill at barEvery ticks and names its box only
// when the whole-pixel painted width changed. A paused board holds the bar
// still.
func (a *App) moveBar() {
	if a.paused || a.flip.bar == nil {
		return
	}

	a.flip.barTick++
	if a.flip.barTick < barEvery {
		return
	}

	a.flip.barTick = 0

	fraction := float64(time.Since(a.lastRotate)) / float64(rotateEvery)
	fraction = min(max(fraction, 0), 1)

	w := a.flip.barFullW * fraction
	if int(w) == int(a.flip.bar.W) {
		return
	}

	a.flip.bar.W = w
	a.page.Invalidate("refresh-fill")
}

// invalidateCells names every tile the flap step changed this frame.
func (a *App) invalidateCells(ids []string) {
	for _, id := range ids {
		a.page.Invalidate(id)
	}
}
