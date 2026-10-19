package csspreview

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestSecondAnimationRebindsFirstAcrossRedraw(t *testing.T) {
	p, d, _ := fixturePage(t)
	startCatalog(t, p, d, "animation")
	first := d.samples[catalogSampleKey("animation")]
	origin := first.origin
	gen := p.Generation()
	startCatalog(t, p, d, "-webkit-animation")
	if p.Generation() <= gen {
		t.Fatal("second click did not redraw")
	}
	if err := d.tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !first.origin.Equal(origin) {
		t.Fatal("first sample origin changed")
	}
	if d.gen != p.Generation() || !planInDisplay(p.Display(), first.plan) {
		t.Fatal("first sample was not rebound to the current display")
	}
	first.plan.step(first.plan.start.Add(first.plan.delay + first.plan.duration/2))
	changed := false
	for _, pair := range first.plan.pairs {
		changed = changed || read(pair.op) != pair.from
	}
	if !changed {
		t.Fatal("first sample could not render its deterministic mid-frame")
	}
}

func planInDisplay(d *ownframe.Display, m motion) bool {
	if d == nil || len(m.pairs) == 0 {
		return false
	}
	for _, pair := range m.pairs {
		found := false
		for i := range d.Ops {
			if pair.op == &d.Ops[i] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
