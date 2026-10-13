package board

import "time"

// wantFrame declares the next frame the board needs, so an idle window wakes
// for it instead of at its keep-alive fallback. While a flap spins the board
// needs every frame. Otherwise the header clock is the only thing that
// changes on its own, so the board asks for its next whole second and the
// refresh bar moves on those wakes. Asking for the bar's 10 Hz cadence while
// nothing else is happening would cost a frame every 100 ms for one fill, so
// the page leaves that cadence to the fast loop, where the bar still steps
// every 100 ms. The window clears the request when it reads it, so the tick
// asks again on every frame.
func (a *App) wantFrame() {
	if a.flip.spinning() {
		a.page.WantFrameIn(0)

		return
	}

	a.page.WantFrameIn(nextSecondIn(time.Now()))
}

// spinning reports whether any tile still has to move.
func (f *flipper) spinning() bool {
	for _, fl := range f.cells {
		if !fl.done {
			return true
		}
	}

	return false
}

// nextSecondIn is the time until the header clock's next whole second, plus a
// cushion so the waking tick reads the new second.
func nextSecondIn(now time.Time) time.Duration {
	return now.Truncate(time.Second).Add(time.Second + 2*time.Millisecond).Sub(now)
}
