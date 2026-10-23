package page

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
)

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

// Stats returns one snapshot of the page counters and the last frame sizes.
// LastDraw is the window's draw time once SetDrawTime has recorded one, and
// the display-list or paint stage of the last Redraw before that.
func (p *Page) Stats() host.Stats {
	ops := 0
	if p.draw.Display != nil {
		ops = len(p.draw.Display.Ops)
	}

	counts := draw.Counts{
		Redraws:         p.stats.redraws,
		Parses:          p.stats.parses,
		Cascades:        p.stats.cascades,
		Layouts:         p.stats.layouts,
		Repaints:        p.stats.repaints,
		LastRedraw:      p.stats.lastRedraw,
		LastDraw:        p.stats.lastDraw,
		FrameDraw:       p.stats.frameDraw,
		LastTemplate:    p.stats.lastTemplate,
		LayoutTime:      p.stats.layoutTime,
		DisplayListTime: p.stats.displayListTime,
		PaintTime:       p.stats.paintTime,
		DirtyOps:        p.stats.dirtyOps,
		DirtyRegions:    p.stats.dirtyRegions,
		ChangedOps:      p.stats.changedOps,
		AllocFrame:      p.stats.allocFrame,
		Reloads:         p.stats.reloads,
		LastReloadError: p.stats.reloadErr,
	}

	return draw.Snapshot(&counts, len(p.boxes), ops)
}
