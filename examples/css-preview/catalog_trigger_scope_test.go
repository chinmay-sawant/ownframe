package main

import (
	"math"
	"testing"
	"time"
)

// TestCatalogTriggerScopeLimitsOwnFrameActions checks that the scope on an
// ancestor decides which listeners hear the event. The target inside the
// scoped section moves; the one outside it does not.
func TestCatalogTriggerScopeLimitsOwnFrameActions(t *testing.T) {
	html := `<style>` + triggerKeyframes +
		`#scope{trigger-scope:--fire}` +
		`#source{width:40px;height:40px;background:#888}` +
		`#target,#outside{width:40px;height:20px;background:#123456;` +
		`animation-name:route;animation-duration:1000ms;animation-fill-mode:both;` +
		`animation-timing-function:linear;animation-trigger:--fire play;}</style>` +
		`<section id="scope"><div id="source"></div><div id="target"></div></section>` +
		`<div id="outside"></div>`
	p := motionPage(t, html, "")
	box, ok := findBox(p.Boxes(), "source")
	if !ok {
		t.Fatal("source box missing")
	}
	if err := p.Click(t.Context(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	in := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel)
	out := xformX(targetOp(t, p, "outside"), p.Probe().Display().PointsPerPixel)
	if in <= 0.1 || math.Abs(out) > .1 {
		t.Fatalf("inside/outside samples = %.2f/%.2f px, want a moving inside and a still outside", in, out)
	}
}
