package draw

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Counts is one page's redraw counters. FrameDraw, once set by the window,
// replaces LastDraw in the snapshot. The other names match host.Stats.
type Counts struct {
	Redraws         uint64
	Parses          uint64
	Cascades        uint64
	Layouts         uint64
	Repaints        uint64
	LastRedraw      time.Duration
	LastDraw        time.Duration
	FrameDraw       time.Duration
	LastTemplate    time.Duration
	LayoutTime      time.Duration
	DisplayListTime time.Duration
	PaintTime       time.Duration
	DirtyOps        int
	DirtyRegions    int
	ChangedOps      int
	AllocFrame      uint64
	Reloads         uint64
	LastReloadError string
}

// Snapshot copies the counters. LastDraw is FrameDraw once that is set,
// and the display-list or paint stage before that. boxes and ops are the
// sizes of the committed frame, which the counters do not hold.
func Snapshot(c *Counts, boxes, ops int) host.Stats {
	lastDraw := c.LastDraw
	if c.FrameDraw > 0 {
		lastDraw = c.FrameDraw
	}

	return host.Stats{
		Redraws:         c.Redraws,
		Parses:          c.Parses,
		Cascades:        c.Cascades,
		Layouts:         c.Layouts,
		Repaints:        c.Repaints,
		Boxes:           boxes,
		Ops:             ops,
		LastRedraw:      c.LastRedraw,
		LastDraw:        lastDraw,
		Reloads:         c.Reloads,
		LastReloadError: c.LastReloadError,
		LastTemplate:    c.LastTemplate,
		LayoutTime:      c.LayoutTime,
		DisplayListTime: c.DisplayListTime,
		PaintTime:       c.PaintTime,
		DirtyOps:        c.DirtyOps,
		DirtyRegions:    c.DirtyRegions,
		ChangedOps:      c.ChangedOps,
		AllocFrame:      c.AllocFrame,
	}
}

// NoteCommit records one committed frame: a layout, a repaint, and the
// time since the display build started.
func NoteCommit(layouts, repaints *uint64, lastDraw *time.Duration, start time.Time) {
	*layouts++
	*repaints++
	*lastDraw = time.Since(start)
}
