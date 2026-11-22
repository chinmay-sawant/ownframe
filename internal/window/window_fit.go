package window

// Fitting a window to its monitor.
//
// A window's size is chosen in the units of the monitor it opens on. Opened on
// a 150 percent panel, a 980 by 756 window is 1470 by 1134 pixels, and it keeps
// that size when it lands on a 1080 pixel high monitor at 100 percent: the
// bottom, or the whole header when the position came from the first monitor,
// hangs off the screen. The fit runs once at startup and once after the window
// settles on a different monitor. It shrinks and moves the window just enough
// to lie inside that monitor's work area. It never runs while the window is
// maximized, minimized, full screen, or being dragged, and it does not
// enlarge a window or move one that already fits.

// fitRect is a window or work area rectangle in desktop pixels.
type fitRect struct{ x, y, w, h int }

// fitSettleUpdates is how many updates a new monitor must stay current before
// the fit runs. A drag across monitors, or the window manager adjusting the
// frame after a DPI change, then settles first.
const fitSettleUpdates = 6

// fitInto returns r shrunk to the size of work if it is larger, and moved so
// it lies inside work. It reports whether the result differs from r. An empty
// work area changes nothing.
func fitInto(r, work fitRect) (fitRect, bool) {
	if work.w <= 0 || work.h <= 0 || r.w <= 0 || r.h <= 0 {
		return r, false
	}

	out := fitRect{w: min(r.w, work.w), h: min(r.h, work.h)}
	out.x = min(max(r.x, work.x), work.x+work.w-out.w)
	out.y = min(max(r.y, work.y), work.y+work.h-out.h)

	return out, out != r
}

// fitProbe is what the platform reports about the window and its monitor.
type fitProbe struct {
	monitor uintptr // identifies the monitor the window is mostly on
	window  fitRect
	work    fitRect
	// skip is true for a window that must be left alone: maximized,
	// minimized, or covering its whole monitor.
	skip bool
}

// windowFit tracks which monitor the window was last fitted for.
type windowFit struct {
	started bool
	monitor uintptr
	pending bool
	wait    int
}

// step runs once per update. probe and apply are the platform hooks; probe
// reports ok false where the platform has no fit (everything but Windows).
func (f *windowFit) step(dragging bool, probe func() (fitProbe, bool), apply func(fitRect) bool) {
	p, ok := probe()
	if !ok {
		return
	}

	switch {
	case !f.started:
		f.started, f.monitor, f.pending, f.wait = true, p.monitor, true, fitSettleUpdates
	case p.monitor != f.monitor:
		f.monitor, f.pending, f.wait = p.monitor, true, fitSettleUpdates
	}

	if !f.pending {
		return
	}

	if dragging || p.skip {
		f.wait = fitSettleUpdates

		return
	}

	if f.wait--; f.wait > 0 {
		return
	}

	f.pending = false

	if r, changed := fitInto(p.window, p.work); changed {
		apply(r)
	}
}
