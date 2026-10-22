package ownframe

import "context"

// SetPicker installs the dialog a file control opens.
func (t Testing) SetPicker(pick func(ctx context.Context, title string) (string, error)) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetPicker(pick)
}

// CancelTouches queues gesture cancellation for the window loop.
func (t Testing) CancelTouches() {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.CancelTouches()
}

// SetTouchZoomAllowed controls two-finger touch zoom.
func (t Testing) SetTouchZoomAllowed(on bool) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetTouchZoomAllowed(on)
}

// TouchZoomAllowed reports whether two-finger touch may zoom.
func (t Testing) TouchZoomAllowed() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.TouchZoomAllowed()
}

// SetTouchScrollOptions tunes finger sensitivity and inertial decay.
func (t Testing) SetTouchScrollOptions(options TouchScrollOptions) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetTouchScrollOptions(options)
}
