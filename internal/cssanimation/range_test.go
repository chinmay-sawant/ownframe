package cssanimation_test

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

func TestOwnFrameAnimationRange(t *testing.T) {
	assertRange(t, `animation-range:25% 75%;`, .5, 5)
}

func TestOwnFrameAnimationRangeStart(t *testing.T) {
	assertRange(t, `animation-range-start:25%;`, .625, 5)
}

func TestOwnFrameAnimationRangeEnd(t *testing.T) {
	assertRange(t, `animation-range-end:75%;`, .375, 5)
}

func TestOwnFrameAnimationRangeUsesHostNamedRanges(t *testing.T) {
	d, _, indices, paints, p, _ := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-timeline:scroll();animation-range:entry 0% exit 100%;animation-fill-mode:both;animation-timing-function:linear;`)
	p.StepWithInput(animation.FrameInput{ScrollProgress: .5, Ranges: map[string]animation.Range{
		"entry": {Start: .2, End: .4}, "exit": {Start: .6, End: .8},
	}})
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
		t.Errorf("named range moved %v px, want 5", got)
	}
}

func assertRange(t *testing.T, property string, progress, want float64) {
	t.Helper()
	d, _, indices, paints, p, _ := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-timeline:scroll();animation-fill-mode:both;animation-timing-function:linear;`+property)
	p.StepWithInput(animation.FrameInput{ScrollProgress: progress})
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-want) > .02 {
		t.Errorf("%s at progress %.2f moved %v px, want %v", property, progress, got, want)
	}
}
