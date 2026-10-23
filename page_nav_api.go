package ownframe

import "context"

// HTML returns the template source at the current history index.
func (p *Page) HTML() string { return innerPage(p).HTML() }

// Load parses html, drops later history, and draws that page.
func (p *Page) Load(ctx context.Context, html string) error {
	return innerPage(p).Load(ctx, html)
}

// Back draws the previous template source.
func (p *Page) Back(ctx context.Context) error {
	return innerPage(p).Back(ctx)
}

// Forward draws the next template source.
func (p *Page) Forward(ctx context.Context) error {
	return innerPage(p).Forward(ctx)
}

// Route maps a data-action value to HTML for Click to load.
func (p *Page) Route(action, html string) error {
	return innerPage(p).Route(action, html)
}
