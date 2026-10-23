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
	// lastTemplate, styleTime, layoutTime, and paintTime split the
	// last Redraw into stages. dirtyOps, dirtyRegions, and changedOps
	// describe its dirty region. allocFrame is the bytes allocated during
	// the last sampled Redraw. Sampling runs about twice a second.
	// Unsampled redraws keep the previous value.
	lastTemplate time.Duration
	styleTime    time.Duration
	layoutTime   time.Duration
	paintTime    time.Duration
	otherTime    time.Duration
	dirtyOps     int
	dirtyRegions int
	dirtyArea    int
	changedOps   int
	allocFrame   uint64
}
