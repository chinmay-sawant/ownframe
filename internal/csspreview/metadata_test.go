package csspreview

import (
	"strings"
	"testing"
)

func TestCombinedFixtureExportsAnimationAndTransitionMetadata(t *testing.T) {
	p, _, _ := fixturePage(t, "all")
	standardTarget, _ := targetByKey("all", catalogSampleKey("animation"))
	standard, _ := findBox(p.Boxes(), standardTarget.id)
	a := standard.Animations[0]
	if a.Name != "standard-glide" || a.DurationMS != 1600 || a.TimingFunction != "cubic-bezier(0.22, 0.75, 0.25, 1)" || a.IterationInfinite || a.IterationCount != 2.5 || a.DelayStartMS != 100 || a.DelayEndMS != 150 || a.Direction != "alternate" || a.FillMode != "both" || a.PlayState != "paused" || a.Composition != "replace" || a.Timeline != "auto" || a.RangeStart != "entry 10%" || a.RangeEnd != "exit 90%" || a.RangeCenter != "normal" || !strings.Contains(a.Trigger, "--standard-start") {
		t.Fatalf("animation metadata lost fields: %+v", a)
	}
	if standard.TriggerScope != "--standard-start" || standard.ViewTransitionName != "standard-chip" || standard.ViewTransitionClass != "animation-demo" || standard.ViewTransitionGroup != "normal" || standard.ViewTransitionScope != "all" {
		t.Fatalf("box metadata lost fields: %+v", standard)
	}
}
