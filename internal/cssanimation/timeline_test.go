package cssanimation_test

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

func TestOwnFrameAnimationTimeline(t *testing.T) {
	cases := []struct {
		timeline string
		input    animation.FrameInput
	}{
		{"scroll()", animation.FrameInput{ScrollProgress: .5}},
		{"view()", animation.FrameInput{ViewProgress: .5}},
		{"--custom", animation.FrameInput{Timelines: map[string]float64{"--custom": .5}}},
	}
	for _, c := range cases {
		d, _, indices, paints, p, _ := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-iteration-count:1;animation-fill-mode:both;animation-timing-function:linear;animation-timeline:`+c.timeline+`;`)
		p.StepWithInput(c.input)
		if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
			t.Errorf("timeline %s moved %v px at halfway, want 5", c.timeline, got)
		}
	}
}
