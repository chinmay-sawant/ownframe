package csspreview

import (
	"context"
	"testing"
)

func TestStartChangesOnlyItsTarget(t *testing.T) {
	for _, kind := range []string{"animation", "transition"} {
		for _, selected := range targets(kind) {
			t.Run(kind+"/"+selected.key, func(t *testing.T) {
				ctx := context.Background()
				p, d, v := fixturePage(t, kind)
				base := map[string][]paint{}
				for _, target := range targets(kind) {
					box, ok := findBox(p.Boxes(), target.id)
					if !ok {
						t.Fatalf("missing target box %q", target.id)
					}
					base[target.key] = capture(p.Display(), box)
					if !hasRecords(p, box, target, kind) {
						t.Fatalf("missing Blinkless records for %q", target.key)
					}
				}
				button, ok := findBox(p.Boxes(), selected.button)
				if !ok {
					t.Fatalf("missing start button %q", selected.button)
				}
				if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
					t.Fatal(err)
				}
				if err := d.tick(ctx); err != nil {
					t.Fatal(err)
				}
				s := d.samples[selected.key]
				if s == nil || len(s.plan.pairs) == 0 {
					t.Fatalf("button %q produced no motion", selected.button)
				}
				s.plan.step(s.plan.start.Add(s.plan.delay + s.plan.duration/2))
				changed := false
				for _, pair := range s.plan.pairs {
					changed = changed || read(pair.op) != pair.from || pair.op.Opacity() != pair.from.a*pair.from.opacity
				}
				if !changed {
					t.Fatalf("target %q did not change", selected.id)
				}
				for _, target := range targets(kind) {
					if target.key == selected.key {
						continue
					}
					box, _ := findBox(p.Boxes(), target.id)
					if !samePaint(base[target.key], capture(p.Display(), box)) {
						t.Errorf("starting %q changed sibling %q", selected.key, target.key)
					}
				}
				if !v.Started[selected.key] {
					t.Fatalf("start key %q was not recorded", selected.key)
				}
			})
		}
	}
}
