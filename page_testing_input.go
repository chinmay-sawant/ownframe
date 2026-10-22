package ownframe

import "context"

// Click hit-tests the last picture and applies the click.
func (page *Page) Click(ctx context.Context, x, y float64) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Click(ctx, x, y)
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
