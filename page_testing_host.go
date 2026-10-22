package ownframe

import (
	"context"
	"time"
)

// PollReload rereads a watched source file.
func (page *Page) PollReload(ctx context.Context) (bool, error) {
	p := innerPage(page)
	if p == nil {
		return false, nil
	}

	return p.PollReload(ctx)
}

// PatchOp mutates one retained operation.
func (page *Page) PatchOp(i int, fn func(o *DisplayOp)) bool {
	p := innerPage(page)
	if p == nil {
		return false
	}

	return p.PatchOp(i, fn)
}

// Tick runs the frame callback registered with SetTick and the caret blink.
func (t Testing) Tick(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Tick(ctx)
}

// SetAfterRedraw registers fn to run after a redraw commits.
func (page *Page) SetAfterRedraw(fn func()) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetAfterRedraw(fn)
}

// BeginRedrawBatch defers draws until EndRedrawBatch.
func (page *Page) BeginRedrawBatch() {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.BeginRedrawBatch()
}

// EndRedrawBatch closes a batch and draws the final state once.
func (page *Page) EndRedrawBatch(ctx context.Context) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.EndRedrawBatch(ctx)
}

// DispatchDeepLink validates a host-delivered URL and calls the app handler.
func (t Testing) DispatchDeepLink(ctx context.Context, target string) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.DispatchDeepLink(ctx, target)
}

// PrepareMobile prepares this page for its native mobile viewport.
func (t Testing) PrepareMobile(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.PrepareMobile(ctx)
}

// SetPerf turns pipeline timing, dirty counters, and alloc tracking on or off.
func (page *Page) SetPerf(on bool) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetPerf(on)
}

// Perf reports whether pipeline timing and dirty counters are recorded.
func (page *Page) Perf() bool {
	p := innerPage(page)
	if p == nil {
		return false
	}

	return p.Perf()
}

// SetDevTools turns the window devtools overlay on or off.
func (page *Page) SetDevTools(on bool) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetDevTools(on)
}

// DevTools reports whether the window devtools overlay starts on.
func (t Testing) DevTools() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.DevTools()
}

// PatchDisplay mutates the retained display and records changed regions.
func (page *Page) PatchDisplay(fn func(d *Display)) DisplayPatch {
	p := innerPage(page)
	if p == nil {
		return DisplayPatch{}
	}

	return p.PatchDisplay(fn)
}

// LastPatch returns the last display patch report.
func (page *Page) LastPatch() DisplayPatch {
	p := innerPage(page)
	if p == nil {
		return DisplayPatch{}
	}

	return p.LastPatch()
}

// SetPicker installs the dialog a file control opens.
func (t Testing) SetPicker(pick func(ctx context.Context, title string) (string, error)) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetPicker(pick)
}

// SetTouchZoomAllowed controls two-finger touch zoom.
func (page *Page) SetTouchZoomAllowed(on bool) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetTouchZoomAllowed(on)
}

// TouchZoomAllowed reports whether two-finger touch may zoom.
func (page *Page) TouchZoomAllowed() bool {
	p := innerPage(page)
	if p == nil {
		return false
	}

	return p.TouchZoomAllowed()
}

// SetTouchScrollOptions tunes finger sensitivity and inertial decay.
func (page *Page) SetTouchScrollOptions(options TouchScrollOptions) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetTouchScrollOptions(options)
}

// SetHotReload turns the file watch on or off.
func (page *Page) SetHotReload(on bool) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetHotReload(on)
}

// Watching reports whether a poll can reread a source file.
func (page *Page) Watching() bool {
	p := innerPage(page)
	if p == nil {
		return false
	}

	return p.Watching()
}

// WantFrameAt declares the next time the page needs the window to draw.
func (page *Page) WantFrameAt(at time.Time) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.WantFrameAt(at)
}

// WantFrameIn declares the next frame in d.
func (page *Page) WantFrameIn(d time.Duration) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.WantFrameIn(d)
}
