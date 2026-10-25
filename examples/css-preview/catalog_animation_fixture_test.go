package main

import (
	"testing"
)

// animationRuntimeCase names one animation property and two values that must
// reach a different painted frame. The fixture pauses the animation at a
// negative delay, so the sampled frame is the same on every run.
type animationRuntimeCase struct {
	property, before, after string
	// base overrides the shared declaration. A timing case uses it to put
	// the sample past the end of the active interval.
	base          string
	compareActive bool
}

// animationBase pauses a one second route at a quarter of its duration. A
// case that tests timing overrides the matching longhand.
const animationBase = `animation-name:route;animation-duration:1000ms;` +
	`animation-delay:-250ms;animation-fill-mode:both;` +
	`animation-timing-function:linear;animation-play-state:paused;`

func animationCSS(c animationRuntimeCase, value string) string {
	base := animationBase
	if c.base != "" {
		base = c.base
	}
	if c.property == "animation" || c.property == "-webkit-animation" {
		base = ""
	}
	return `#target{width:40px;height:20px;background-color:rgb(20,40,60);` +
		base + c.property + `:` + value + `;}` + keyframeCSS
}

const keyframeCSS = `@keyframes route{from{transform:translateX(0px)}to{transform:translateX(100px)}}
@keyframes returnRoute{from{transform:translateX(0px)}to{transform:translateX(-100px)}}
@-webkit-keyframes route{from{transform:translateX(0px)}to{transform:translateX(100px)}}
@-webkit-keyframes returnRoute{from{transform:translateX(0px)}to{transform:translateX(-100px)}}`

// sampleAnimation builds the page, draws it, ticks the automatic driver once,
// and returns the target's sampled paint and whether the driver has work.
func sampleAnimation(t *testing.T, c animationRuntimeCase, value string) (paint, bool) {
	t.Helper()
	p := motionPage(t, `<div id="target"></div>`, animationCSS(c, value))
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	paints := targetPaints(t, p, "target")
	if len(paints) == 0 {
		t.Fatal("target painted nothing")
	}
	return paints[0], p.Probe().Motion().Active > 0
}
