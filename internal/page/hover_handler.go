package page

import "context"

// HoverHandler may paint a hover change directly into the retained display.
// Returning true skips CSS relayout; the page invalidates both boxes. Return
// false for any change that needs layout or ordinary CSS state handling.
// The handler must not change box geometry. An error leaves hover unchanged.
type HoverHandler func(ctx context.Context, previous, next Box) (handled bool, err error)

func (p *Page) paintHover(ctx context.Context, id string) (bool, error) {
	if p.handlers.Hover == nil || p.display == nil {
		return false, nil
	}
	return p.handlers.Hover(ctx, p.boxByID(p.hoverID()), p.boxByID(id))
}
