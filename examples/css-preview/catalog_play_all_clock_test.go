package main

import (
	"context"
	"testing"
)

// TestPlayAllTakesItsStartTimeAtTheBind checks that a Play all run leaves
// the start time to the sample. A clock stamped in the click handler is
// spent by the redraw that follows the click, and the shortest sample would
// be over before it began.
func TestPlayAllTakesItsStartTimeAtTheBind(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	all, ok := findBox(p.Boxes(), playAllID)
	if !ok {
		t.Fatal("the play all button is missing")
	}
	if err := p.Click(ctx, all.X+all.W/2, all.Y+all.H/2); err != nil {
		t.Fatal(err)
	}
	for _, row := range v.Catalog {
		s := d.samples[row.SampleKey]
		if s == nil {
			t.Fatalf("%s has no sample", row.Name)
		}
		if !s.origin.IsZero() {
			t.Fatalf("%s took its start time in the click", row.Name)
		}
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
	if len(d.pending) != 0 {
		t.Fatalf("pending = %d, want an empty queue", len(d.pending))
	}
}

// TestPlayAllTwiceKeepsOneQueueEntryPerRow checks that the pending queue
// holds one entry per row, so a second Play all press does not grow it.
func TestPlayAllTwiceKeepsOneQueueEntryPerRow(t *testing.T) {
	ctx := context.Background()
	p, d, _ := fixturePage(t)
	all, ok := findBox(p.Boxes(), playAllID)
	if !ok {
		t.Fatal("the play all button is missing")
	}
	for i := 0; i < 2; i++ {
		if err := p.Click(ctx, all.X+all.W/2, all.Y+all.H/2); err != nil {
			t.Fatal(err)
		}
	}
	if len(d.pending) != len(d.v.Catalog) {
		t.Fatalf("pending = %d, want one entry per row", len(d.pending))
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if len(d.pending) != 0 {
		t.Fatalf("pending = %d, want an empty queue", len(d.pending))
	}
}
