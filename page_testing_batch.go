package ownframe

import "context"

// SetAfterRedraw registers fn to run after a redraw commits.
func (t Testing) SetAfterRedraw(fn func()) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetAfterRedraw(fn)
}

// BeginRedrawBatch defers draws until EndRedrawBatch.
func (t Testing) BeginRedrawBatch() {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.BeginRedrawBatch()
}

// EndRedrawBatch closes a batch and draws the final state once.
func (t Testing) EndRedrawBatch(ctx context.Context) error {
	p := innerPage(t.page)
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

// ViewLocked reports that the window should scale this page to the screen.
func (t Testing) ViewLocked() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.ViewLocked()
}
