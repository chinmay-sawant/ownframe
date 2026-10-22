package ownframe

import "context"

// Click hit-tests the last picture and applies the click.
func (t Testing) Click(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Click(ctx, x, y)
}

// Hover sets the hovered element from a point.
func (t Testing) Hover(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Hover(ctx, x, y)
}

// Press sets the pressed element from a point.
func (t Testing) Press(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Press(ctx, x, y)
}

// Release clears the pressed element.
func (t Testing) Release(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Release(ctx)
}

// Type appends text to the focused field.
func (t Testing) Type(ctx context.Context, text string) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Type(ctx, text)
}
