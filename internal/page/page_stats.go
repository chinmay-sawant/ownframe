package page

import "time"

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
	// describe its dirty region. allocFrame is the bytes allocated during
	// the last sampled Redraw. Sampling runs about twice a second.
	// Unsampled redraws keep the previous value.
	lastTemplate    time.Duration
	layoutTime      time.Duration
	displayListTime time.Duration
	paintTime       time.Duration
	dirtyOps        int
	dirtyRegions    int
	changedOps      int
	allocFrame      uint64
}
