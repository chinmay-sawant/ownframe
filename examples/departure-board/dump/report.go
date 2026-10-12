package dump

import (
	"runtime"

	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// Rect is a JSON-friendly CSS-pixel rectangle.
type Rect struct{ X0, Y0, X1, Y1 int }

// Frame is one measured frame after a Tick.
type Frame struct {
	Kind        string // "full-replay", "full-buffer", "rect", or "blit"
	TickNS      int64
	Ops         int
	ReplayedOps int // ops meeting the dirty rect, approximate for kind "rect"
	Rect        Rect
	FullArea    int // content area in CSS pixels
}

// Report is one scenario's measurement.
type Report struct {
	Scenario     string
	Ops          int
	Boxes        int
	RedrawNS     int64
	Frames       []Frame
	Kinds        map[string]int
	TickP50NS    int64
	TickP95NS    int64
	TickMaxNS    int64
	RectMaxArea  int
	RectMeanArea int
	AllocsFrame  float64
}

// summarize fills the histogram, the tick percentiles, the rect areas, and
// the allocation rate, then records the final op and box counts.
func (rep *Report) summarize(app *board.App, before, after runtime.MemStats) {
	ticks := make([]int64, len(rep.Frames))
	areaSum, areaN := 0, 0
	for i, f := range rep.Frames {
		ticks[i] = f.TickNS
		rep.Kinds[f.Kind]++
		if a := rectArea(f.Rect); a > 0 {
			areaSum += a
			areaN++
			if a > rep.RectMaxArea {
				rep.RectMaxArea = a
			}
		}
	}
	if areaN > 0 {
		rep.RectMeanArea = areaSum / areaN
	}
	rep.TickP50NS = percentile(ticks, 50)
	rep.TickP95NS = percentile(ticks, 95)
	rep.TickMaxNS = percentile(ticks, 100)
	if n := len(rep.Frames); n > 0 {
		rep.AllocsFrame = float64(after.TotalAlloc-before.TotalAlloc) / float64(n)
	}

	p := app.Page()
	rep.Boxes = len(p.Boxes())
	if d := p.Display(); d != nil {
		rep.Ops = len(d.Ops)
	}
}

// rectArea is the pixel area of r, or 0 when r is empty or unset.
func rectArea(r Rect) int {
	w, h := r.X1-r.X0, r.Y1-r.Y0
	if w <= 0 || h <= 0 {
		return 0
	}
	return w * h
}
