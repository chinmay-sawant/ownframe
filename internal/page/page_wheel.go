package page

import "context"

// Wheel offers one wheel event to the page before the window pans the page.
// The handler returns true when it consumed the event. A page without a
// handler leaves the event to the window.
func (p *Page) Wheel(ctx context.Context, x, y, dx, dy float64) (bool, error) {
	if p == nil || p.handlers.Wheel == nil {
		return false, nil
	}

	if err := useContext(ctx); err != nil {
		return false, err
	}

	return p.handlers.Wheel(ctx, x, y, dx, dy)
}
