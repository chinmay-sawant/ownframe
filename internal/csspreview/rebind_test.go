package csspreview

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestSecondAnimationRebindsFirstAcrossRedraw(t *testing.T) {
	ctx := context.Background()
	p, d, _ := fixturePage(t, "animation")
	startAnimation(t, p, d, "standard")
	first := d.samples["standard"]
	origin := first.origin
	gen := p.Generation()
	startAnimation(t, p, d, "webkit")
	if p.Generation() <= gen {
		t.Fatal("second click did not redraw")
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if !first.origin.Equal(origin) {
		t.Fatal("first sample origin changed")
	}
	if d.gen != p.Generation() || !planInDisplay(p.Display(), first.plan) {
		t.Fatal("first sample was not rebound to the current display")
	}
	first.plan.step(first.plan.start.Add(first.plan.duration / 2))
	changed := false
	for _, pair := range first.plan.pairs {
		changed = changed || read(pair.op) != pair.from
	}
	if !changed {
		t.Fatal("first sample could not render its deterministic mid-frame")
	}
}

func startAnimation(t *testing.T, p *ownframe.Page, d *demo, key string) {
	t.Helper()
	target, _ := targetByKey("animation", key)
	button, ok := findBox(p.Boxes(), target.button)
	if !ok {
		t.Fatalf("missing button %q", target.button)
	}
	if err := p.Click(context.Background(), button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
	if err := d.tick(context.Background()); err != nil {
		t.Fatal(err)
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
