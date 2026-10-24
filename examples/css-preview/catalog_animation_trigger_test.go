package main

import (
	"math"
	"testing"
	"time"
)

const triggerKeyframes = `@keyframes route{from{transform:translateX(0px)}to{transform:translateX(100px)}}`

// TestCatalogAnimationTriggerChangesOwnFrameAfterHostEvent drives the page's
// own trigger dispatch. A press inside the scope plays the target; the
// release runs the reset action and returns it to the first keyframe.
func TestCatalogAnimationTriggerChangesOwnFrameAfterHostEvent(t *testing.T) {
	html := `<style>` + triggerKeyframes +
		`#scope{trigger-scope:--fire}` +
		`#target{width:40px;height:20px;background:#123456;` +
		`animation-name:route;animation-duration:1000ms;animation-fill-mode:both;` +
		`animation-timing-function:linear;animation-trigger:--fire play reset;}</style>` +
		`<section id="scope"><button id="source">go</button><div id="target"></div></section>`
	p := motionPage(t, html, "")
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel); math.Abs(got) > .1 {
		t.Fatalf("pending animation moved to %.2f px, want the declared base at 0", got)
	}
	box, ok := findBox(p.Boxes(), "source")
	if !ok {
		t.Fatal("source box missing")
	}
	x, y := box.X+box.W/2, box.Y+box.H/2
	if err := p.Edit().Press(t.Context(), x, y); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	played := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel)
	if played <= 0.1 || played >= 110 {
		t.Fatalf("triggered sample = %.2f px, want it past the base and before the end", played)
	}
	if err := p.Edit().Release(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel); math.Abs(got) > .1 {
		t.Fatalf("reset trigger sample = %.2f px, want the first keyframe at 0", got)
	}
}
