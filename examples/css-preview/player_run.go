package main

import (
	"context"
)

// wake arms the tick when the app has work for the frame loop.
func (d *demo) wake() {
	if walkActive() {
		return
	}
	if len(d.pending) > 0 || d.remeasure {
		d.p.SetTick(d.tick)
	}
}

// endRun closes a Play all run whose previews have all finished.
func (d *demo) endRun(ctx context.Context, running bool) {
	if !d.allPlaying || d.playAllAt.IsZero() || running || len(d.pending) > 0 {
		return
	}
	d.allPlaying = false
	d.setReadout(ctx, "All previews played.")
}

// setReadout puts the readout text into the parsed tree. The template keeps
// the text it last rendered, so the tree carries the current one.
func (d *demo) setReadout(ctx context.Context, text string) {
	d.readout = text
	if !d.patchReadout(ctx, text) {
		// The engine refused the readout, so the page has to draw again
		// and the template renders it.
		d.v.Readout = text
	}
}
