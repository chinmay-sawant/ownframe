package ownframe

import "context"

// SetPicker installs the dialog a file control opens.
func (p *Page) SetPicker(pick func(ctx context.Context, title string) (string, error)) {
	innerPage(p).SetPicker(pick)
}

// CancelTouches queues gesture cancellation for the window loop.
func (p *Page) CancelTouches() { innerPage(p).CancelTouches() }

// SetTouchZoomAllowed controls two-finger touch zoom.
func (p *Page) SetTouchZoomAllowed(on bool) {
	innerPage(p).SetTouchZoomAllowed(on)
}

// TouchZoomAllowed reports whether two-finger touch may zoom.
func (p *Page) TouchZoomAllowed() bool { return innerPage(p).TouchZoomAllowed() }

// SetTouchScrollOptions tunes finger sensitivity and inertial decay.
func (p *Page) SetTouchScrollOptions(options TouchScrollOptions) {
	innerPage(p).SetTouchScrollOptions(options)
}

// SetAfterRedraw registers fn to run after a redraw commits.
func (p *Page) SetAfterRedraw(fn func()) { innerPage(p).SetAfterRedraw(fn) }

// BeginRedrawBatch defers draws until EndRedrawBatch.
func (p *Page) BeginRedrawBatch() { innerPage(p).BeginRedrawBatch() }

// EndRedrawBatch closes a batch and draws the final state once.
func (p *Page) EndRedrawBatch(ctx context.Context) error {
	return innerPage(p).EndRedrawBatch(ctx)
}

// DispatchDeepLink validates a host-delivered URL and calls the app handler.
func (p *Page) DispatchDeepLink(ctx context.Context, target string) error {
	return innerPage(p).DispatchDeepLink(ctx, target)
}

// ViewLocked reports that the window should scale this page to the screen.
func (p *Page) ViewLocked() bool { return innerPage(p).ViewLocked() }
