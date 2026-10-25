package main

import (
	"testing"
)

// TestAnimationTickMarksPartialDamage checks that one automatic sample marks
// the card, not the whole page, and that the tick opts into the frame dirty
// path.
func TestAnimationTickMarksPartialDamage(t *testing.T) {
	p := motionPage(t, `<div id="target"></div><div id="still"></div>`,
		`@keyframes route{from{transform:translateX(0px)}to{transform:translateX(20px)}}
		#target{width:40px;height:20px;background:#2459d3;animation:route 1s linear infinite}
		#still{width:40px;height:20px;margin-top:80px;background:#999}`)
	p.Probe().TakeDirtyRects()
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !p.Probe().FrameDirty() {
		t.Fatal("animation tick did not opt into the frame-dirty path")
	}
	rects, ok := p.Probe().TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatal("animation tick marked no damage")
	}
	for _, rect := range rects {
		if rect.Dx() > 100 || rect.Dy() > 100 {
			t.Fatalf("tick damage %v is the whole page, want the animated element", rect)
		}
	}
}

// TestFinishedAnimationStopsAskingForFrames checks that the automatic driver
// parks once its sample is over. A paused sample is sampled once and then
// skipped, so a still page stops requesting frames.
func TestFinishedAnimationStopsAskingForFrames(t *testing.T) {
	p := motionPage(t, `<div id="target"></div>`,
		`@keyframes route{from{transform:translateX(0px)}to{transform:translateX(20px)}}
		#target{width:20px;height:20px;background:#123456;
		animation:route 20ms linear 1 normal both paused;}`)
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if p.Probe().Motion().Active != 0 {
		t.Fatal("a paused sample kept the driver active")
	}
	if !p.Testing().NextFrameAt().IsZero() {
		t.Fatal("a paused sample asked for another frame")
	}
}
