package main

import (
	"testing"
	"time"
)

// TestFinishedForwardFillSurvivesAnotherStart checks that a later click's
// redraw rebuilds the page but restores the finished animation's fill.
func TestFinishedForwardFillSurvivesAnotherStart(t *testing.T) {
	p, d, _ := fixturePage(t)
	startCatalog(t, p, d, "animation-fill-mode")
	first := d.samples[catalogSampleKey("animation-fill-mode")]
	at := first.origin.Add(first.plan.delay + first.plan.duration + time.Second)
	first.plan.step(at)
	finish := read(first.plan.pairs[0].op)
	first.origin = time.Now().Add(-5 * time.Second)
	first.active = false

	startCatalog(t, p, d, "-webkit-animation")
	if got := read(first.plan.pairs[0].op); got != finish {
		t.Fatalf("fill after redraw = %+v, want %+v", got, finish)
	}
}
