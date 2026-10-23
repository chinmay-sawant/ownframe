package host

import "time"

// Stats is one snapshot of the page counters an inspector shows. The zero
// value is a page that has drawn nothing.
type Stats struct {
	Redraws    uint64
	Parses     uint64
	Cascades   uint64
	Layouts    uint64
	Repaints   uint64
	Boxes      int
	Ops        int
	LastRedraw time.Duration
	LastDraw   time.Duration
	// LastTemplate, LayoutTime, DisplayListTime, and PaintTime split the
	// last Redraw into template, style, display-list, and paint stages.
	LastTemplate    time.Duration
	LayoutTime      time.Duration
	DisplayListTime time.Duration
	PaintTime       time.Duration
	// DirtyOps counts display ops meeting the dirty rect, DirtyRegions is
	// 1 when the last Redraw dirtied a region, and ChangedOps counts the
	// ops the display diff found changed.
	DirtyOps     int
	DirtyRegions int
	ChangedOps   int
	// AllocFrame is the bytes allocated during the last sampled Redraw.
	// Sampling runs about twice a second. Unsampled redraws keep the
	// previous value.
	AllocFrame uint64
	Reloads    uint64
	// LastReloadError is the most recent failed hot reload, empty when none.
	LastReloadError string
}

// Inspector is a screen with a devtools overlay. The window draws the
// overlay; the page reports the stats because only the page can measure a
// parse.
type Inspector interface {
	DevTools() bool
	SetDevTools(on bool)
	Stats() Stats
}
