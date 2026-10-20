package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestCatalogStartDispatchesAnimationTriggers(t *testing.T) {
	for _, name := range []string{"animation", "animation-trigger", "trigger-scope"} {
		t.Run(name, func(t *testing.T) {
			p, d, _ := fixturePage(t)
			startCatalog(t, p, d, name)
			s := d.samples[catalogSampleKey(name)]
			if s == nil || s.plan.animation == nil {
				t.Fatal("Start did not bind an animation player")
			}
			active := s.plan.step(s.plan.start.Add(s.plan.delay + s.plan.duration/2))
			changed := false
			for _, pair := range s.plan.pairs {
				changed = changed || read(pair.op) != pair.from
			}
			if !active || !changed {
				t.Fatalf("Start did not dispatch the trigger: active=%v changed=%v", active, changed)
			}
		})
	}
}

func startCatalog(t *testing.T, p *ownframe.Page, d *demo, name string) {
	t.Helper()
	var row catalogRow
	for _, candidate := range d.v.Catalog {
		if candidate.Name == name {
			row = candidate
			break
		}
	}
	if row.Name == "" {
		t.Fatalf("missing property %q", name)
	}
	button, ok := findBox(p.Boxes(), row.Button)
	if !ok {
		t.Fatalf("missing catalog button %q", row.Button)
	}
	if err := p.Click(context.Background(), button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
	if err := d.tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}
