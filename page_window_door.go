package ownframe

import "context"

// Window reads and sets reload, the animation and transition switches, perf,
// devtools, and touch policy.
type Window struct{ page *Page }

// Window returns reload, the animation and transition switches, perf, devtools,
// and touch policy for this page.
func (p *Page) Window() Window { return Window{page: p} }

// PollReload rereads a watched source file.
func (w Window) PollReload(ctx context.Context) (bool, error) {
	p := innerPage(w.page)
	if p == nil {
		return false, nil
	}

	return p.PollReload(ctx)
}

// SetAnimations turns the automatic CSS motion driver on or off. It is on by
// default; turning it off releases the bound players at once.
func (w Window) SetAnimations(on bool) {
	p := innerPage(w.page)
	if p == nil {
		return
	}
	p.SetAnimations(on)
}

// Animations reports whether the automatic CSS motion driver is on.
func (w Window) Animations() bool {
	p := innerPage(w.page)
	if p == nil {
		return false
	}

	return p.Animations()
}

// SetTransitions turns the automatic CSS transition driver on or off. It is on
// by default. Turning it off releases the live renderers without writing, so a
// transition in flight freezes at the paint it had already presented; turning
// it back on lets the next style change capture. The animation switch is
// independent.
func (w Window) SetTransitions(on bool) {
	p := innerPage(w.page)
	if p == nil {
		return
	}
	p.SetTransitions(on)
}

// Transitions reports whether the automatic CSS transition driver is on.
func (w Window) Transitions() bool {
	p := innerPage(w.page)
	if p == nil {
		return false
	}

	return p.Transitions()
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
