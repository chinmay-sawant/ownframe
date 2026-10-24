package main

import (
	"context"
	"time"
)

// reset clears the walk state on the game goroutine. The HTTP /reset handler
// only sets resetReq; step performs the view, manual sample, and page reset.
func (w *walker) reset(ctx context.Context) {
	w.v.Started = map[string]bool{}
	w.v.Restart = map[string]bool{}
	w.v.Restarting = map[string]bool{}
	w.v.Selected = ""
	w.v.Readout = ""
	// The readout and the card classes are template data, so the copy in
	// the demo has to go too.
	w.d.readout = ""
	w.d.manual = map[string]*manualSample{}
	w.d.startedAt = time.Time{}
	w.d.startedSeq = 0
	w.next, w.scrollFor = -1, -1
	w.aim = [3]int{-1, 0, 0}
	w.p.ScrollTo(0, 0)
	w.v.applyRows()
	w.p.SetData(w.v)
	_ = w.p.Redraw(ctx)
}
