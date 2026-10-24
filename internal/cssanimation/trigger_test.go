package cssanimation_test

import (
	"testing"
	"time"
)

func TestOwnFrameTriggeredAnimationWaitsForAction(t *testing.T) {
	d, _, indices, paints, p, start := runtimePlayer(t, `animation-name:ten,twenty;animation-duration:1s;animation-timing-function:linear;animation-composition:add,add;animation-fill-mode:both;animation-trigger:none,--fire play;`)
	p.Step(start.Add(time.Second))
	if got := movedX(d, indices, paints) / d.PointsPerPixel; got < 9.98 || got > 10.02 {
		t.Fatalf("pending triggered effect changed paint: %v px", got)
	}
	if err := p.TriggerEffect(1, "play", start.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	p.Step(start.Add(1500 * time.Millisecond))
	// A transform-list keyframe writes one matrix per effect, so the activated
	// effect's 10px replaces the first effect's 10px instead of summing with it.
	if got := movedX(d, indices, paints) / d.PointsPerPixel; got < 9.98 || got > 10.02 {
		t.Fatalf("activated second effect moved %v px, want 10", got)
	}
}
