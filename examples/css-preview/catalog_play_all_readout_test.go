package main

import (
	"context"
	"strings"
	"testing"
)

// TestPlayAllKeepsItsReadout checks that the readout of a Play all run
// survives a rebuild. The readout lives in the parsed tree, and a rebuild
// replaces the tree, so the readout has to be patched back.
func TestPlayAllKeepsItsReadout(t *testing.T) {
	ctx := context.Background()
	p, d, _ := fixturePage(t)
	all, ok := findBox(p.Boxes(), playAllID)
	if !ok {
		t.Fatal("the play all button is missing")
	}
	if err := p.Testing().Click(ctx, all.X+all.W/2, all.Y+all.H/2); err != nil {
		t.Fatal(err)
	}
	if d.readout != "All previews are playing." {
		t.Fatalf("readout = %q, want the run readout", d.readout)
	}
	// A changed template value rebuilds the tree.
	d.v.Readout = "a rebuilt readout"
	p.SetData(d.v)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	out, ok := findBox(p.Boxes(), readoutID)
	if !ok {
		t.Fatal("the readout is missing")
	}
	if !strings.Contains(out.Text, "All previews are playing.") {
		t.Fatalf("the readout reads %q", out.Text)
	}
}
