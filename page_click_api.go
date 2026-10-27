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
