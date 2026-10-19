package main

import "testing"

func TestWebKitAnimationCardsExposeTheirPropertyValues(t *testing.T) {
	p, _, _ := fixturePage(t)
	want := map[string]struct {
		name      string
		duration  float64
		delay     float64
		direction string
		fill      string
		count     float64
	}{
		"-webkit-animation-delay":           {name: "wk-delay", duration: 1200, delay: 350},
		"-webkit-animation-direction":       {name: "wk-direction", direction: "reverse"},
		"-webkit-animation-duration":        {name: "wk-duration", duration: 1200},
		"-webkit-animation-fill-mode":       {name: "wk-fill", fill: "forwards", count: 1},
		"-webkit-animation-iteration-count": {name: "wk-iterations", count: 3},
	}
	for name, expected := range want {
		target, _ := targetByKey(catalogSampleKey(name))
		box, _ := findBox(p.Boxes(), target.id)
		if len(box.Animations) != 1 {
			t.Errorf("%s animation metadata: %+v", name, box.Animations)
			continue
		}
		actual := box.Animations[0]
		if actual.Name != expected.name ||
			(expected.duration > 0 && actual.DurationMS != expected.duration) ||
			(expected.delay > 0 && actual.DelayStartMS != expected.delay) ||
			(expected.direction != "" && actual.Direction != expected.direction) ||
			(expected.fill != "" && actual.FillMode != expected.fill) ||
			(expected.count > 0 && actual.IterationCount != expected.count) {
			t.Errorf("%s metadata = %+v, want %+v", name, actual, expected)
		}
	}
}
