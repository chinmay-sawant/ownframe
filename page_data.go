package ownframe

import "context"

// SetData stores the value the template prints on the next Redraw.
func (p *Page) SetData(data any) { innerPage(p).SetData(data) }

// Handle registers the input functions. A later call replaces the whole set.
func (p *Page) Handle(h Handlers) { innerPage(p).Handle(h) }

// AddHandlers merges only non-nil callbacks into the existing set.
func (p *Page) AddHandlers(h Handlers) { innerPage(p).AddHandlers(h) }

// SetTheme stores an extra stylesheet the next Redraw applies.
func (p *Page) SetTheme(source string) error {
	return innerPage(p).SetTheme(source)
}

// SetTick registers fn to run once per frame before the window draws.
func (p *Page) SetTick(fn func(ctx context.Context) error) {
	innerPage(p).SetTick(fn)
}

// SetImage registers encoded image bytes for one source name.
func (p *Page) SetImage(src string, data []byte) {
	innerPage(p).SetImage(src, data)
}
