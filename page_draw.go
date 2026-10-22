package ownframe

import "context"

// Redraw renders the current size.
func (p *Page) Redraw(ctx context.Context) error {
	return innerPage(p).Redraw(ctx)
}

// RedrawIfChanged renders unless the last committed frame is unchanged.
func (p *Page) RedrawIfChanged(ctx context.Context) error {
	return innerPage(p).RedrawIfChanged(ctx)
}

// Title returns the window title.
func (p *Page) Title() string { return innerPage(p).Title() }

// Generation increases by one on every successful Redraw.
func (p *Page) Generation() uint64 { return innerPage(p).Generation() }
