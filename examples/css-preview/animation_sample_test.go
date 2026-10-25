package main

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// The keyframes mirror the catalog's standard-glide card. A paused negative
// delay fixes the sampled frame at 50%, so the assertion does not depend on
// when the tick lands.
const standardGlideCSS = `@keyframes standard-glide {
	0% { transform: translateX(0px); opacity: 1; }
	50% { transform: translateX(90px); opacity: 0.72; background-color: #e27a27; }
	100% { transform: translateX(180px); opacity: 1; background-color: #2459d3; }
}
#target { width: 50px; height: 27px; background-color: #2459d3;
	animation: standard-glide 1600ms linear -800ms 1 normal both paused; }`

func TestAnimationSamplesTransformOpacityAndColor(t *testing.T) {
	p := motionPage(t, `<div id="target"></div>`, standardGlideCSS)
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	op := targetOp(t, p, "target")
	moved := xformX(op, p.Probe().Display().PointsPerPixel)
	if math.Abs(moved-90) > .01 {
		t.Errorf("x = %v, want 90px translation", moved)
	}
	if math.Abs(op.Opacity()-.72) > .01 || math.Abs(op.R-226.0/255) > .01 {
		t.Errorf("midframe paint opacity=%v red=%v", op.Opacity(), op.R)
	}
	if op.Kind != ownframe.DisplayOpFillRect {
		t.Errorf("sampled op kind = %v, want a fill", op.Kind)
	}
}
