package ownframe

import "context"

// LongPress fires a held press at a point.
func (p *Page) LongPress(ctx context.Context, x, y float64) (bool, error) {
	return innerPage(p).LongPress(ctx, x, y)
}

// Swipe delivers a swipe delta to the handler.
func (p *Page) Swipe(ctx context.Context, dx, dy float64) error {
	return innerPage(p).Swipe(ctx, dx, dy)
}

// BeginDrag starts a pointer drag.
func (p *Page) BeginDrag(ctx context.Context, x, y float64) (bool, error) {
	return innerPage(p).BeginDrag(ctx, x, y)
}

// MoveDrag continues a pointer drag.
func (p *Page) MoveDrag(ctx context.Context, x, y float64) error {
	return innerPage(p).MoveDrag(ctx, x, y)
}

// EndDrag finishes a pointer drag.
func (p *Page) EndDrag(ctx context.Context) error {
	return innerPage(p).EndDrag(ctx)
}
