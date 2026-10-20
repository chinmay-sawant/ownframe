package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstrigger"
)

func TestCatalogTriggerScopeLimitsOwnFrameActions(t *testing.T) {
	scopeName := "--fire"
	style := `#scope{trigger-scope:` + scopeName + `}#target,#outside{width:40px;height:20px;background:#123456;animation-name:route;animation-duration:1s;animation-fill-mode:both;animation-timing-function:linear;animation-trigger:--fire play;}` + keyframeCSS
	html := `<section id="scope"><button id="source"></button><div id="target"></div></section><div id="outside"></div>`
	d, _, _, _ := motionFixture(t, html, style)
	start := time.Unix(100, 0)
	inside, target, inOps, inBase := bindCatalogTarget(t, d, "target", start)
	outside, other, outOps, outBase := bindCatalogTarget(t, d, "outside", start)
	inside.Step(start.Add(time.Second))
	outside.Step(start.Add(time.Second))
	source := catalogBoxIndex(d.Boxes, "source")
	err := csstrigger.Dispatch(d.Boxes, csstrigger.Event{Name: "--fire", Source: source, Active: true}, map[int]*animation.Player{target: inside, other: outside}, start.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	inside.Step(start.Add(1500 * time.Millisecond))
	outside.Step(start.Add(1500 * time.Millisecond))
	in := (cssmotion.Read(&d.Ops[inOps[0]]).X - inBase[0].X) / d.PointsPerPixel
	out := cssmotion.Read(&d.Ops[outOps[0]]).X - outBase[0].X
	if in < 49.9 || in > 50.1 || out != 0 {
		t.Fatalf("inside/outside samples = %.2f/%.2f px, want 50/0", in, out/d.PointsPerPixel)
	}
}
