package ownframe

import (
	"context"
	"time"
)

// Testing is for test files. Example programs call Page methods directly.
// The methods here simulate input or read host probes that an app does not call.
type Testing struct{ page *Page }

// Testing returns helpers that belong in test files.
func (p *Page) Testing() Testing { return Testing{page: p} }

// Tick runs the frame callback registered with SetTick and the caret blink.
func (t Testing) Tick(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Tick(ctx)
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

// DevTools reports whether the window devtools overlay starts on.
func (t Testing) DevTools() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.DevTools()
}

// SetPicker installs the dialog a file control opens.
func (t Testing) SetPicker(pick func(ctx context.Context, title string) (string, error)) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetPicker(pick)
}

// LongPress fires a held press at a point.
func (t Testing) LongPress(ctx context.Context, x, y float64) (bool, error) {
	p := innerPage(t.page)
	if p == nil {
		return false, nil
	}

	return p.LongPress(ctx, x, y)
}

// Swipe delivers a swipe delta to the handler.
func (t Testing) Swipe(ctx context.Context, dx, dy float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Swipe(ctx, dx, dy)
}

// BeginDrag starts a pointer drag.
func (t Testing) BeginDrag(ctx context.Context, x, y float64) (bool, error) {
	p := innerPage(t.page)
	if p == nil {
		return false, nil
	}

	return p.BeginDrag(ctx, x, y)
}

// MoveDrag continues a pointer drag.
func (t Testing) MoveDrag(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.MoveDrag(ctx, x, y)
}

// EndDrag finishes a pointer drag.
func (t Testing) EndDrag(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.EndDrag(ctx)
}

// IMEReplace replaces a byte range in the focused value.
func (t Testing) IMEReplace(ctx context.Context, start, end int, insert string, caret int) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.IMEReplace(ctx, start, end, insert, caret)
}

// IMEContext returns the focused field's box and caret text.
func (t Testing) IMEContext() (Box, int, string, string, bool) {
	p := innerPage(t.page)
	if p == nil {
		return Box{}, 0, "", "", false
	}

	return p.IMEContext()
}

// SelectAt focuses the control under the point and places the caret.
func (t Testing) SelectAt(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.SelectAt(ctx, x, y)
}

// PNGWithContext threads ctx to the on-demand render.
func (t Testing) PNGWithContext(ctx context.Context) []byte {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.PNGWithContext(ctx)
}

// ChangedProperties returns changed-property metadata surfaced to the page.
func (t Testing) ChangedProperties() []string {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.ChangedProperties()
}

// ClearChangedProperties clears surfaced metadata.
func (t Testing) ClearChangedProperties() {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.ClearChangedProperties()
}

// TakeChangedProperties returns a snapshot and drains local metadata.
func (t Testing) TakeChangedProperties() []string {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.TakeChangedProperties()
}

// TakeScroll returns the pending scroll request and clears it.
func (t Testing) TakeScroll() (ScrollRequest, bool) {
	p := innerPage(t.page)
	if p == nil {
		return ScrollRequest{}, false
	}
	req, ok := p.TakeScroll()
	return ScrollRequest{X: req.X, Y: req.Y, Absolute: req.Absolute}, ok
}

// StepScrollWindow prepares visible rows and reports whether they changed.
func (t Testing) StepScrollWindow() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.StepScrollWindow()
}

// NextFrameAt returns the deadline set by WantFrameAt and clears it.
func (t Testing) NextFrameAt() time.Time {
	p := innerPage(t.page)
	if p == nil {
		return time.Time{}
	}

	return p.NextFrameAt()
}
