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
	// LastTemplate, StyleTime, LayoutTime, and PaintTime split the
	// last Redraw into template execution, CSS styling, placement/list emission,
	// and software paint. OtherTime covers form sync, dirty diff, callbacks,
	// and remaining bookkeeping. Stages describe this call, including cache hits.
	LastTemplate time.Duration
	StyleTime    time.Duration
	LayoutTime   time.Duration
	PaintTime    time.Duration
	OtherTime    time.Duration
	// DirtyOps counts each op intersecting any changed region once. DirtyRegions
	// and DirtyArea describe the unpadded display-diff rectangles in CSS pixels.
	// The host may pad, merge, or tile these regions before repainting.
	DirtyOps     int
	DirtyRegions int
	DirtyArea    int
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
