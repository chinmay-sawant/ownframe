package main

import (
	"context"
	"time"
)

// reset clears the walk state on the game goroutine. The HTTP /reset handler
// only sets resetReq; step performs the view, sample, and page reset.
func (w *walker) reset(ctx context.Context) {
	w.v.Started = map[string]bool{}
	w.v.Selected = ""
	w.v.Readout = ""
	// The readout lives in the parsed tree, so the copy in the demo has to
	// go too. The after-redraw hook patches it back into every rebuilt
	// tree, and a stale copy would keep the last row's readout on screen.
	w.d.readout = ""
	w.d.samples = map[string]*sample{}
	w.d.pending = nil
	w.d.allPlaying = false
	w.d.playAllAt = time.Time{}
	w.next, w.scrollFor = -1, -1
	w.aim = [3]int{-1, 0, 0}
	w.p.ScrollTo(0, 0)
	w.d.gen = w.p.Probe().Generation()
	w.p.SetData(w.v)
	_ = w.p.Redraw(ctx)
}
