package ownframe

import "context"

// Window reads and sets reload, perf, devtools, and touch policy.
type Window struct{ page *Page }

// Window returns reload, perf, devtools, and touch policy for this page.
func (p *Page) Window() Window { return Window{page: p} }

// PollReload rereads a watched source file.
func (w Window) PollReload(ctx context.Context) (bool, error) {
	p := innerPage(w.page)
	if p == nil {
		return false, nil
	}

	return p.PollReload(ctx)
}

// SetPerf turns pipeline timing, dirty counters, and alloc tracking on or off.
func (w Window) SetPerf(on bool) {
	p := innerPage(w.page)
	if p == nil {
		return
	}
	p.SetPerf(on)
}

// Perf reports whether pipeline timing and dirty counters are recorded.
func (w Window) Perf() bool {
	p := innerPage(w.page)
	if p == nil {
		return false
	}

	return p.Perf()
}

// SetDevTools turns the window devtools overlay on or off.
func (w Window) SetDevTools(on bool) {
	p := innerPage(w.page)
	if p == nil {
		return
	}
	p.SetDevTools(on)
}

// SetTouchZoomAllowed controls two-finger touch zoom.
func (w Window) SetTouchZoomAllowed(on bool) {
	p := innerPage(w.page)
	if p == nil {
		return
	}
	p.SetTouchZoomAllowed(on)
}

// TouchZoomAllowed reports whether two-finger touch may zoom.
func (w Window) TouchZoomAllowed() bool {
	p := innerPage(w.page)
	if p == nil {
		return false
	}

	return p.TouchZoomAllowed()
}

// SetTouchScrollOptions tunes finger sensitivity and inertial decay.
func (w Window) SetTouchScrollOptions(options TouchScrollOptions) {
	p := innerPage(w.page)
	if p == nil {
		return
	}
	p.SetTouchScrollOptions(options)
}

// SetHotReload turns the file watch on or off.
func (w Window) SetHotReload(on bool) {
	p := innerPage(w.page)
	if p == nil {
		return
	}
	p.SetHotReload(on)
}

// Watching reports whether a poll can reread a source file.
func (w Window) Watching() bool {
	p := innerPage(w.page)
	if p == nil {
		return false
	}

	return p.Watching()
}

// DevTools reports whether the page requests the window devtools overlay.
func (w Window) DevTools() bool {
	p := innerPage(w.page)
	return p != nil && p.DevTools()
}
