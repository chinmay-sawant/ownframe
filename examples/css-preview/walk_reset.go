package main

import "context"

// reset clears the walk state on the game goroutine. The HTTP /reset handler
// only sets resetReq; step performs the view, sample, and page reset.
func (w *walker) reset() {
	w.v.Started = map[string]bool{}
	w.v.Selected = ""
	w.v.Readout = ""
	w.d.samples = map[string]*sample{}
	w.d.pending = nil
	w.next, w.scrollFor = -1, -1
	w.aim = [3]int{-1, 0, 0}
	w.p.ScrollTo(0, 0)
	w.d.gen = w.p.Generation()
	w.p.SetData(w.v)
	_ = w.p.Redraw(context.Background())
}
