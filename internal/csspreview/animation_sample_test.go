package csspreview

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestAnimationSamplesTransformOpacityAndColor(t *testing.T) {
	p, d, _ := fixturePage(t, "all")
	startCatalog(t, p, d, "animation")
	m := d.samples[catalogSampleKey("animation")].plan
	m.step(m.start.Add(m.delay + m.duration/2))
	found := false
	for _, pair := range m.pairs {
		if pair.op.Kind != ownframe.DisplayOpFillRect || pair.from.b < .5 {
			continue
		}
		found = true
		if math.Abs(pair.op.X-pair.from.x-90*p.Display().PointsPerPixel) > .01 {
			t.Errorf("x = %v, want 90px translation", pair.op.X-pair.from.x)
		}
		if math.Abs(pair.op.Opacity()-.72) > .01 || math.Abs(pair.op.R-226.0/255) > .01 {
			t.Errorf("midframe paint opacity=%v red=%v", pair.op.Opacity(), pair.op.R)
		}
	}
	if !found {
		t.Fatal("animation keyframes did not update the target fill")
	}
	progress, active, apply := m.position(m.start.Add(m.delay + 5*m.duration/2))
	if !active || !apply || math.Abs(progress-.5) > .001 {
		t.Fatalf("final fractional iteration = %v, active=%v apply=%v", progress, active, apply)
	}
	_, active, apply = m.position(m.start.Add(m.delay + 5*m.duration/2 + m.endDelay + time.Millisecond))
	if active || !apply {
		t.Fatalf("end delay state active=%v apply=%v", active, apply)
	}
}
