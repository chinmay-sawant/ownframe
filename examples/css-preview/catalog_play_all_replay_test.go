package main

import (
	"context"
	"testing"
)

// TestStartReplaysAfterPlayAll checks that a Start button still replays its
// own preview after the Play all button started every card.
func TestStartReplaysAfterPlayAll(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	all, ok := findBox(p.Boxes(), playAllID)
	if !ok {
		t.Fatal("the play all button is missing")
	}
	if err := p.Click(ctx, all.X+all.W/2, all.Y+all.H/2); err != nil {
		t.Fatal(err)
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}

	row := v.Catalog[33]
	s := d.samples[row.SampleKey]
	if s == nil {
		t.Fatalf("%s has no sample", row.Name)
	}
	first := s.origin

	button, ok := findBox(p.Boxes(), row.Button)
	if !ok {
		t.Fatalf("missing %s", row.Button)
	}
	if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if !s.origin.After(first) {
		t.Fatalf("%s did not restart", row.Name)
	}
	s.plan.step(s.origin.Add(s.plan.duration / 2))
	got := read(s.plan.pairs[0].op)
	if samePaintValue(got, s.plan.pairs[0].to) {
		t.Fatalf("%s did not replay", row.Name)
	}
}
