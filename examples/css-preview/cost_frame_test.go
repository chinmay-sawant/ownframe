package main

import "testing"

// costFrame is what one frame of the real window cost.
type costFrame struct {
	draw   float64 // the window's draw time, ms
	tick   float64 // the page's tick, ms
	alloc  float64 // the bytes the frame allocated, KB
	gap    float64 // the wall time between two frames, ms
	active int     // the samples running in the frame
}

// costValues reads one field of the frames.
func costValues(frames []costFrame, pick func(costFrame) float64) []float64 {
	out := make([]float64, 0, len(frames))
	for _, f := range frames {
		out = append(out, pick(f))
	}

	return out
}

// costActives lists the running sample count of every frame.
func costActives(frames []costFrame) []int {
	out := make([]int, 0, len(frames))
	for _, f := range frames {
		out = append(out, f.active)
	}

	return out
}

// costDraws lists the window's draw time of every frame.
func costDraws(frames []costFrame) []float64 {
	out := make([]float64, 0, len(frames))
	for _, f := range frames {
		out = append(out, f.draw)
	}

	return out
}

// costReportFrames prints the per-frame numbers of one run.
func costReportFrames(tb testing.TB, label string, frames []costFrame) {
	tb.Helper()
	costReport(tb, label+" draw/frame", costValues(frames,
		func(f costFrame) float64 { return f.draw }), "ms")
	costReport(tb, label+" tick/frame", costValues(frames,
		func(f costFrame) float64 { return f.tick }), "ms")
	costReport(tb, label+" alloc/frame", costValues(frames,
		func(f costFrame) float64 { return f.alloc }), "KB")
	costReport(tb, label+" frame period", costValues(frames,
		func(f costFrame) float64 { return f.gap }), "ms")
}
