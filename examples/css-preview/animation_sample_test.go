package main

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestAnimationSamplesTransformOpacityAndColor(t *testing.T) {
	p, d, _ := fixturePage(t)
	startCatalog(t, p, d, "animation")
	m := d.samples[catalogSampleKey("animation")].plan
	m.step(m.start.Add(m.delay + m.duration/2))
	found := false
	for _, pair := range m.pairs {
		if pair.op.Kind != ownframe.DisplayOpFillRect || pair.from.B < .5 {
			continue
		}
		found = true
		if math.Abs(pair.op.X-pair.from.X-90*p.Probe().Display().PointsPerPixel) > .01 {
			t.Errorf("x = %v, want 90px translation", pair.op.X-pair.from.X)
		}
		if math.Abs(pair.op.Opacity()-.72) > .01 || math.Abs(pair.op.R-226.0/255) > .01 {
			t.Errorf("midframe paint opacity=%v red=%v", pair.op.Opacity(), pair.op.R)
		}
	}
	if !found {
		t.Fatal("animation keyframes did not update the target fill")
	}
}
