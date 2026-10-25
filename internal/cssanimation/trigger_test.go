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
	// Both effects are additive, so the activated twenty effect (10px at its
	// midpoint) concatenates after the finished ten effect's 10px.
	if got := movedX(d, indices, paints) / d.PointsPerPixel; got < 19.98 || got > 20.02 {
		t.Fatalf("activated second effect moved %v px, want 20", got)
	}
}
