package page

import "context"

// Redraw renders the current size: a display list for a replayable page,
// a bitmap otherwise. Parsed trees and sheets are reused. Stage timing and
// allocs track only with Perf on.
func (p *Page) Redraw(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	track, m0, start := p.beginRedraw()
	defer p.finishRedraw(track, m0, start)
	executed, err := p.execTemplate(track)
	if err != nil {
		return err
	}
	p.source = p.syncForm(executed)
	p.png = nil
	state := p.renderState()
	styled, err := p.styleStaged(ctx, track, p.source, state)
	if err != nil {
		return err
	}
	return p.drawStaged(ctx, track, styled, state, executed)
}
