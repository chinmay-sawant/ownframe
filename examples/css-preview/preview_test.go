package main

import (
	"context"
	"testing"
)

func TestStartChangesOnlyItsTarget(t *testing.T) {
	for _, name := range []string{"animation", "transition-property"} {
		t.Run(name, func(t *testing.T) {
			p, d, v := fixturePage(t)
			selected, ok := targetByKey(catalogSampleKey(name))
			if !ok {
				t.Fatalf("missing sample for %q", name)
			}
			base := make(map[string][]paint, len(targets()))
			for _, item := range targets() {
				box, found := findBox(p.Boxes(), item.id)
				if !found {
					t.Fatalf("missing target box %q", item.id)
				}
				base[item.key] = capture(p.Testing().Display(), box)
			}
			startCatalog(t, p, d, name)
			sample := d.samples[selected.key]
			if sample == nil || len(sample.plan.pairs) == 0 {
				t.Fatalf("button %q produced no motion", selected.button)
			}
			sample.plan.step(sample.plan.start.Add(sample.plan.delay + sample.plan.duration/2))
			changed := false
			for _, pair := range sample.plan.pairs {
				changed = changed || read(pair.op) != pair.from
			}
			if !changed {
				t.Fatalf("target %q did not change", selected.id)
			}
			for _, item := range targets() {
				if item.key == selected.key {
					continue
				}
				box, _ := findBox(p.Boxes(), item.id)
				if !samePaint(base[item.key], capture(p.Testing().Display(), box)) {
					t.Errorf("starting %q changed sibling %q", selected.key, item.key)
				}
			}
			if !v.Started[selected.key] {
				t.Fatalf("start key %q was not recorded", selected.key)
			}
			if err := d.tick(context.Background()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
