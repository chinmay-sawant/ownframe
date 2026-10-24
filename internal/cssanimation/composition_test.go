package cssanimation_test

import (
	"math"
	"testing"
)

func TestOwnFrameAnimationComposition(t *testing.T) {
	base := `animation-name:ten,twenty;animation-duration:1s;animation-timing-function:linear;animation-fill-mode:both;`
	for _, c := range []struct {
		mode string
		want float64
	}{
		// A transform-list keyframe samples one matrix, so the later effect's
		// matrix replaces the earlier one. Concatenating additive transform
		// lists is a composition phase, not this one.
		{"add,add", 20},
		{"replace,replace", 20},
	} {
		d, _, indices, paints, p, start := runtimePlayer(t, base+`animation-composition:`+c.mode+`;`)
		if p.EffectCount() != 2 {
			t.Fatalf("effects = %d, want 2", p.EffectCount())
		}
		p.Step(start.Add(1_000_000_000))
		got := movedX(d, indices, paints) / d.PointsPerPixel
		if math.Abs(got-c.want) > .02 {
			t.Errorf("composition %s moved %v px, want %v", c.mode, got, c.want)
		}
	}
	d, _, indices, paints, p, start := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-iteration-count:2;animation-composition:accumulate;animation-fill-mode:both;animation-timing-function:linear;`)
	p.Step(start.Add(1500_000_000))
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
		t.Errorf("accumulated animation moved %v px, want current effect value 5", got)
	}
}
