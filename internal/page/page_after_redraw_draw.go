package page

import "context"

// afterRedrawDraw serves the redraw the callback asked for. The draw runs
// after the callback returned, so the commit is never re-entered, and the
// callback runs again in it, so state reapplied to a rebuilt tree lands. One
// draw is enough: a callback that always asks for a draw cannot loop.
func (p *Page) afterRedrawDraw(ctx context.Context) error {
	if !p.afterRedraw.ask {
		return nil
	}

	p.afterRedraw.ask = false
	if p.afterRedraw.served {
		return nil
	}

	p.afterRedraw.served = true
	defer func() { p.afterRedraw.served = false }()

	return p.redraw(ctx, false)
}
