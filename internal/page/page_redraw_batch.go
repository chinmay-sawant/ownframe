package page

import "context"

// BeginRedrawBatch defers page draws until EndRedrawBatch. The window wraps a
// press, focus, click, and release in one batch so the final state draws once.
//
// A redraw that a failed draw left queued stays queued, so the flags survive
// into the next batch and it commits the redraw.
func (p *Page) BeginRedrawBatch() {
	p.redrawBatch++
}

// EndRedrawBatch closes a batch and draws the final state once. An inner batch
// leaves the draw to the outermost one. A plain Redraw in the batch forces the
// final draw; only RedrawIfChanged calls let the final draw skip unchanged
// state.
//
// A draw that fails leaves the redraw queued, so the next batch commits it
// and the committed frame catches up with the state the batch changed.
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

	if err := p.redraw(ctx, !force); err != nil {
		p.redrawPending = true
		p.redrawForce = force

		return err
	}

	return nil
}
