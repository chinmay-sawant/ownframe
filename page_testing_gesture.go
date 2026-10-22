package ownframe

import "context"

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
