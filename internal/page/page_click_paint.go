package page

import "context"

// clickPainted runs the paint-only click handler. It reports whether the
// handler painted the frame itself. A true result clears the pending and
// force flags of the redraw batch, so the batch waits for no draw that
// will not come, and the next real redraw still happens.
func (p *Page) clickPainted(ctx context.Context, box Box) (bool, error) {
	if p.handlers.ClickPaint == nil {
		return false, nil
	}

	painted, err := p.handlers.ClickPaint(ctx, box)
	if err != nil {
		return false, err
	}
	if painted {
		p.paintedFrame()
	}

	return painted, nil
}

// paintedFrame records that the frame is on screen already. The pending and
// force flags of the redraw batch go, so the batch does not wait for a draw
// the handler already made.
func (p *Page) paintedFrame() {
	p.redrawPending = false
	p.redrawForce = false
}
