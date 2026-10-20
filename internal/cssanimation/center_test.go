package cssanimation

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

func TestOwnFrameAnimationRangeCenter(t *testing.T) {
	d, _, indices, paints, p, _ := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-timeline:scroll();animation-range-center:25%;animation-fill-mode:both;animation-timing-function:linear;`)
	p.StepWithInput(animation.FrameInput{ScrollProgress: .25})
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
		t.Errorf("center point moved %v px, want 5", got)
	}
	d, _, indices, paints, p, _ = runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-timeline:scroll();animation-range-center:source 50%;animation-fill-mode:both;animation-timing-function:linear;`)
	p.StepWithInput(animation.FrameInput{ScrollProgress: .3, SourceProgress: .1, TargetProgress: .5})
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
		t.Errorf("source/target center moved %v px, want 5", got)
	}
}
