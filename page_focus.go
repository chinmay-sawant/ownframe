package ownframe

import "context"

// Focus focuses the control id and draws. An empty id blurs.
func (p *Page) Focus(ctx context.Context, id string) error {
	return innerPage(p).Focus(ctx, id)
}

// FocusID returns the focused control id, or "" when none is focused.
func (p *Page) FocusID() string { return innerPage(p).FocusID() }

// FocusNext focuses the next control in document order and wraps.
func (p *Page) FocusNext(ctx context.Context) error {
	return innerPage(p).FocusNext(ctx)
}

// FocusPrev focuses the previous control in document order and wraps.
func (p *Page) FocusPrev(ctx context.Context) error {
	return innerPage(p).FocusPrev(ctx)
}
