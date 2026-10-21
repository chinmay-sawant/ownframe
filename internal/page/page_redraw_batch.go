package page

import "context"

// BeginRedrawBatch defers page draws until EndRedrawBatch. The window wraps a
// press, focus, click, and release in one batch so the final state draws once.
func (p *Page) BeginRedrawBatch() {
	if p.redrawBatch == 0 {
		p.redrawPending = false
		p.redrawForce = false
	}
	p.redrawBatch++
}

// EndRedrawBatch closes a batch and draws the final state once. An inner batch
// leaves the draw to the outermost one. A plain Redraw in the batch forces the
// final draw; only RedrawIfChanged calls let the final draw skip unchanged
// state.
func (p *Page) EndRedrawBatch(ctx context.Context) error {
	if p.redrawBatch == 0 {
		return nil
	}
	p.redrawBatch--
	if p.redrawBatch > 0 || !p.redrawPending {
		return nil
	}

	force := p.redrawForce
	p.redrawPending = false
	p.redrawForce = false

	return p.redraw(ctx, !force)
}
