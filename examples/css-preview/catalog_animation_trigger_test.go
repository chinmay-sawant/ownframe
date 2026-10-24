package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/csstrigger"
)

func TestCatalogAnimationTriggerChangesOwnFrameAfterHostEvent(t *testing.T) {
	action := "play reset"
	style := `#target{width:40px;height:20px;background:#123456;animation-name:route;animation-duration:1s;animation-fill-mode:both;animation-timing-function:linear;animation-trigger:--fire ` + action + `;}` + keyframeCSS + `@keyframes route{from{transform:translateX(10px)}to{transform:translateX(110px)}}`
	d, _, _, _ := motionFixture(t, `<button id="source">go</button><div id="target"></div>`, style)
	start := time.Unix(100, 0)
	p, target, ops, base := bindCatalogTarget(t, d, "target", start)
	p.Step(start.Add(time.Second))
	if got := xformDeltaX(&d.Ops[ops[0]], base[0]); got != 0 {
		t.Fatalf("pending animation moved %v points, want 0", got)
	}
	source := catalogBoxIndex(d.Boxes, "source")
	if source < 0 {
		t.Fatal("source box missing")
	}
	if err := csstrigger.Dispatch(d.Boxes, csstrigger.Event{Name: "--fire", Source: source, Active: true}, map[int]*animation.Player{target: p}, start.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	p.Step(start.Add(1500 * time.Millisecond))
	if got := xformDeltaX(&d.Ops[ops[0]], base[0]) / d.PointsPerPixel; got < 59.9 || got > 60.1 {
		t.Fatalf("triggered sample = %.2f px, want 60", got)
	}
	if err := csstrigger.Dispatch(d.Boxes, csstrigger.Event{Name: "--fire", Source: source, Active: false}, map[int]*animation.Player{target: p}, start.Add(1500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	p.Step(start.Add(1600 * time.Millisecond))
	if got := xformDeltaX(&d.Ops[ops[0]], base[0]) / d.PointsPerPixel; got < 9.9 || got > 10.1 {
		t.Fatalf("reset trigger sample = %.2f px, want first keyframe at 10", got)
	}
}
