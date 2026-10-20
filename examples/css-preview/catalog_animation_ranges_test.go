package main

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

func TestCatalogAnimationRangeReachesOwnFrameFrame(t *testing.T) {
	assertRangeSample(t, "animation-range", "entry 10px exit 90px", animation.FrameInput{
		ScrollProgress: .5,
		RangeEdges:     map[string]float64{"entry 10px": .25, "exit 90px": .75},
	})
}
func TestCatalogAnimationRangeStartReachesOwnFrameFrame(t *testing.T) {
	assertRangeSample(t, "animation-range-start", "25%", animation.FrameInput{ScrollProgress: .625})
}
func TestCatalogAnimationRangeEndReachesOwnFrameFrame(t *testing.T) {
	assertRangeSample(t, "animation-range-end", "75%", animation.FrameInput{ScrollProgress: .375})
}
func TestCatalogAnimationRangeCenterReachesOwnFrameFrame(t *testing.T) {
	center := "source 50%"
	got := sampleMotion(t, `animation-name:route;animation-duration:1s;animation-timeline:scroll();animation-fill-mode:both;animation-timing-function:linear;animation-range-center:`+center+`;`, animation.FrameInput{ScrollProgress: .3, SourceProgress: .1, TargetProgress: .5})
	if math.Abs(got-50) > .1 {
		t.Fatalf("animation-range-center:%s sampled %.2f px, want 50", center, got)
	}
}

func assertRangeSample(t *testing.T, property, value string, input animation.FrameInput) {
	t.Helper()
	css := `animation-name:route;animation-duration:1s;animation-timeline:scroll();animation-fill-mode:both;animation-timing-function:linear;` + property + `:` + value + `;`
	got := sampleMotion(t, css, input)
	if math.Abs(got-50) > .1 {
		t.Fatalf("%s:%s sampled %.2f px, want 50", property, value, got)
	}
}
