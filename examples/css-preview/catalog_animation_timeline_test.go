package main

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

func TestCatalogAnimationTimelineSamplesHostProgress(t *testing.T) {
	for _, c := range []struct {
		name, timeline string
		input          animation.FrameInput
	}{{"scroll", "scroll()", animation.FrameInput{ScrollProgress: .25}},
		{"view", "view()", animation.FrameInput{ViewProgress: .75}},
		{"named", "--read", animation.FrameInput{Timelines: map[string]float64{"--read": .5}}}} {
		got := sampleMotion(t, `animation-name:route;animation-duration:1s;animation-fill-mode:both;animation-timing-function:linear;animation-timeline:`+c.timeline+`;`, c.input)
		want := map[string]float64{"scroll": 25, "view": 75, "named": 50}[c.name]
		if math.Abs(got-want) > .1 {
			t.Errorf("%s timeline sample = %.2f px, want %.2f", c.name, got, want)
		}
	}
}
