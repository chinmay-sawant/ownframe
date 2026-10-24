package main

import (
	"math"
	"testing"
)

// TestFinishedForwardFillSurvivesAnotherRedraw checks that a finished
// animation keeps its filled pose when the page is drawn again.
func TestFinishedForwardFillSurvivesAnotherRedraw(t *testing.T) {
	p := motionPage(t, `<div id="target"></div>`,
		`@keyframes route{from{transform:translateX(0px)}to{transform:translateX(60px)}}
		#target{width:20px;height:20px;background:#123456;
		animation:route 20ms linear -50ms 1 normal forwards paused;}`)
	// A paused fill-mode sample needs its first tick to reach the filled pose.
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	finish := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel)
	if math.Abs(finish-60) > .1 {
		t.Fatalf("fill pose = %.2f px, want 60", finish)
	}
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel); math.Abs(got-finish) > .1 {
		t.Fatalf("fill after redraw = %.2f px, want %.2f", got, finish)
	}
}
