package ownframe

import "context"

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

