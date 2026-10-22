package page

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// pageStats counts the stages of every Redraw. The devtools overlay reads
// them through Stats.
type pageStats struct {
	redraws    uint64
	parses     uint64
	cascades   uint64
	layouts    uint64
	repaints   uint64
	reloads    uint64
	reloadErr  string
	lastRedraw time.Duration
	lastDraw   time.Duration
	frameDraw  time.Duration
	// lastTemplate, layoutTime, displayListTime, and paintTime split the
	// last Redraw into stages. dirtyOps, dirtyRegions, and changedOps
	// describe its dirty region; allocFrame counts its allocations.
	lastTemplate    time.Duration
	layoutTime      time.Duration
	displayListTime time.Duration
	paintTime       time.Duration
	dirtyOps        int
	dirtyRegions    int
	changedOps      int
	allocFrame      uint64
}

// Stats returns one snapshot of the page counters and the last frame sizes.
// LastDraw is the window's draw time once SetDrawTime has recorded one, and
// the display-list or paint stage of the last Redraw before that.
func (p *Page) Stats() host.Stats {
	ops := 0
	if p.draw.Display != nil {
		ops = len(p.draw.Display.Ops)
	}

	lastDraw := p.stats.lastDraw
	if p.stats.frameDraw > 0 {
		lastDraw = p.stats.frameDraw
	}

	return host.Stats{
		Redraws:         p.stats.redraws,
		Parses:          p.stats.parses,
		Cascades:        p.stats.cascades,
		Layouts:         p.stats.layouts,
		Repaints:        p.stats.repaints,
		Boxes:           len(p.boxes),
		Ops:             ops,
		LastRedraw:      p.stats.lastRedraw,
		LastDraw:        lastDraw,
		Reloads:         p.stats.reloads,
		LastReloadError: p.stats.reloadErr,
		LastTemplate:    p.stats.lastTemplate,
		LayoutTime:      p.stats.layoutTime,
		DisplayListTime: p.stats.displayListTime,
		PaintTime:       p.stats.paintTime,
		DirtyOps:        p.stats.dirtyOps,
		DirtyRegions:    p.stats.dirtyRegions,
		ChangedOps:      p.stats.changedOps,
		AllocFrame:      p.stats.allocFrame,
	}
}
