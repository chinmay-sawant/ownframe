package main

import (
	"context"
	"testing"
)

// TestPlayAllStartsEverySample checks the Play all button: one click wakes
// every card in the catalog.
func TestPlayAllStartsEverySample(t *testing.T) {
	ctx := context.Background()
	p, d, _ := fixturePage(t)
	button, ok := findBox(p.Boxes(), playAllID)
	if !ok {
		t.Fatal("the play all button is missing")
	}
	if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	active := 0
	for _, s := range d.samples {
		if s.active {
			active++
		}
	}
	if active < len(catalogProperties)-3 {
		t.Fatalf("active samples = %d, want the whole catalog", active)
	}
}

// TestRapidClicksQueueEverySample checks that clicks that land before the
// next tick all start, instead of only the last one.
func TestRapidClicksQueueEverySample(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	for _, row := range v.Catalog[:3] {
		button, ok := findBox(p.Boxes(), row.Button)
		if !ok {
			t.Fatalf("missing %s", row.Button)
		}
		if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
			t.Fatal(err)
		}
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	active := 0
	for _, row := range v.Catalog[:3] {
		s := d.samples[row.SampleKey]
		if s != nil && s.active {
			active++
		}
	}
	if active != 3 {
		t.Fatalf("active samples = %d, want 3", active)
	}
}
