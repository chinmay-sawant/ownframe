package main

import (
	"context"
	"time"
)

// escortEnterSlowStripAfter is the delay after the load-entrance flip before
// the slow 3s bar class is stripped, restoring the 0.9s toggle duration.
const escortEnterSlowStripAfter = 3200 * time.Millisecond

// escortMotion keeps the frame loop hot for dur after a range switch so CSS
// transitions and keyframes sample to completion without pointer input. A
// mutating tick marks every frame dirty (the idle policy needs five clean
// frames to park) and arming a tick flips FramePolling on, enabling the
// keep-alive fallback. It also finishes a pending load entrance: flipping
// enterBars here (before the switch's own refresh commits) keeps a switch
// during load from stranding zero-scale bars, and stripping enterSlow at
// disarm restores the 0.9s toggle duration. The escort always disarms itself;
// SetTick(nil) replaces it, so a later tick owner wins.
func (a *app) escortMotion(dur time.Duration) {
	if a.page == nil {
		return
	}
	if a.enterBars {
		a.enterBars = false
	}
	until := time.Now().Add(dur)
	a.page.SetTick(func(ctx context.Context) error {
		if time.Now().After(until) {
			if a.enterSlow {
				a.enterSlow = false
				if err := a.refresh(ctx); err != nil {
					return err
				}
			}
			// Hand a still-pending toast back to its own ticker so the
			// 5s expiry still runs; the toast ticker clears itself on
			// expiry. Otherwise disarm.
			if a.toastOpen && !a.toastUntil.IsZero() {
				a.page.SetTick(a.tickToast)
			} else {
				a.page.SetTick(nil)
			}
			return nil
		}
		if a.toastOpen {
			if err := a.closeToastExpired(ctx); err != nil {
				return err
			}
		}
		a.page.Probe().UseFrameDirty()
		return nil
	})
}

// escortEnter plays the load entrance hands-free: the growth chart mounts at
// scaleY(0) the first time the Overview tab renders; a tick later it flips to
// real values (a normal 0→target transition capture over the 3s slow class),
// and 3.2s after the flip the slow class is stripped so later toggles use the
// 0.9s transition. The flip waits for tab == "overview" so the rise plays
// when the chart is actually on screen, not on app launch while another tab
// shows, and for one presented zero frame first (the window batches one
// frame's refreshes into a single commit, so flipping in the same frame would
// swallow the zero state and the run would have no start paint). Pre-flip
// ticks stay clean so an unvisited chart costs only the keep-alive poll.
// Then it disarms, handing a pending toast back to its ticker like
// escortMotion does.
func (a *app) escortEnter() {
	if a.page == nil {
		return
	}
	var flippedAt time.Time
	a.page.SetTick(func(ctx context.Context) error {
		if a.enterBars {
			if a.tab != "overview" {
				return nil
			}
			if !a.enterSeen {
				// Let the zero-scale commit present on its own frame
				// first: the window batches one frame's refreshes into a
				// single commit, so flipping in the same frame would
				// swallow the zero state and the run would have no start
				// paint. Mark dirty to keep the loop hot meanwhile.
				a.enterSeen = true
				a.page.Probe().UseFrameDirty()
				return nil
			}
			a.enterBars = false
			flippedAt = time.Now()
			return a.refresh(ctx)
		}
		if !flippedAt.IsZero() && time.Now().After(flippedAt.Add(escortEnterSlowStripAfter)) {
			a.enterSlow = false
			if err := a.refresh(ctx); err != nil {
				return err
			}
			// Hand a still-pending toast back to its own ticker so the
			// 5s expiry still runs; the toast ticker clears itself on
			// expiry. Otherwise disarm.
			if a.toastOpen && !a.toastUntil.IsZero() {
				a.page.SetTick(a.tickToast)
			} else {
				a.page.SetTick(nil)
			}
			return nil
		}
		if a.toastOpen {
			if err := a.closeToastExpired(ctx); err != nil {
				return err
			}
		}
		a.page.Probe().UseFrameDirty()
		return nil
	})
}

// closeToastExpired mirrors tickToast in topbar_panels.go: while the toast is
// still within its deadline it does nothing; once past the deadline it
// clears toastOpen/toastUntil and refreshes (refresh sees a closed toast and
// arms no new tick, so the escort tick survives — toastUntil was already
// zeroed before refresh, so ensureToastTick takes no action).
func (a *app) closeToastExpired(ctx context.Context) error {
	if a.toastOpen && time.Now().Before(a.toastUntil) {
		return nil
	}
	a.toastOpen = false
	a.toastUntil = time.Time{}
	return a.refresh(ctx)
}
