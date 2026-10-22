package page

import "context"

// Redraw renders the current size: a display list for a replayable page,
// a bitmap otherwise. Parsed trees and sheets are reused. Stage timing and
// allocs track only with Perf on. One Page has one owner: never call Redraw
// concurrently on the same Page; keep Tick, MarkRect, and Redraw on the
// page loop the way Run and Serve each own their Page.
func (p *Page) Redraw(ctx context.Context) error {
	return p.redraw(ctx, false)
}

// RedrawIfChanged renders unless the executed source, viewport, focus,
// hover, theme, and images all match the last committed frame. It is for
// callers that redraw after every event and would otherwise lay out the
// same frame again. A skipped call counts no layout or repaint.
func (p *Page) RedrawIfChanged(ctx context.Context) error {
	return p.redraw(ctx, true)
}

func (p *Page) redraw(ctx context.Context, skipSame bool) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if p.afterRedraw.busy {
		// The after-redraw callback asked for a draw. The commit that runs
		// the callback must not start another one inside it, so the request
		// waits for the commit to finish.
		p.afterRedraw.ask = true

		return nil
	}
	if p.redrawBatch > 0 {
		p.redrawPending = true
		if !skipSame {
			p.redrawForce = true
		}
		return nil
	}
	track, m0, start := p.beginRedraw()
	defer p.finishRedraw(track, m0, start)
	executed, err := p.execTemplate(track)
	if err != nil {
		return err
	}
	p.source = p.syncForm(executed)
	state := p.renderState()
	if skipSame && p.unchanged(state) {
		return nil
	}
	p.draw.DropPNG()
	styled, err := p.styleStaged(ctx, track, p.source, state)
	if err != nil {
		return err
	}
	if err := p.drawStaged(ctx, track, styled, state, executed); err != nil {
		return err
	}

	return p.afterRedrawDraw(ctx)
}
