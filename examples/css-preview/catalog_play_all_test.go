package main

import (
	"strings"
	"testing"
	"time"
)

// TestPlayAllMovesEverySample checks that Play all gives every card the same
// result as its own Start button: one sample moves each target off its
// declared paint.
func TestPlayAllMovesEverySample(t *testing.T) {
	p, _, v := fixturePage(t)
	// Settle the paused cards before the comparison.
	tickPage(t, p)
	base := make(map[string][]paint, len(targets()))
	for _, item := range targets() {
		base[item.key] = targetPaints(t, p, item.id)
	}
	clickID(t, p, playAllID)
	time.Sleep(600 * time.Millisecond)
	tickPage(t, p)

	for _, row := range v.Catalog {
		if samePaint(base[row.SampleKey], targetPaints(t, p, row.TargetID)) {
			t.Errorf("%s did not move off its starting paint", row.Name)
		}
	}
}

// TestPlayAllKeepsItsReadout checks that the readout is template data and
// survives a rebuild without a patch hook.
func TestPlayAllKeepsItsReadout(t *testing.T) {
	p, d, _ := fixturePage(t)
	clickID(t, p, playAllID)
	if d.readout != "All previews are playing." {
		t.Fatalf("readout = %q, want the run readout", d.readout)
	}
	// A rebuild renders the view again; the readout is template data.
	d.v.applyRows()
	p.SetData(d.v)
	if err := p.Redraw(t.Context()); err != nil {
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

// TestStartReplaysAfterPlayAll checks that a Start button restarts a card the
// Play all button already started. The restart is the class variant flip.
func TestStartReplaysAfterPlayAll(t *testing.T) {
	p, d, v := fixturePage(t)
	clickID(t, p, playAllID)
	time.Sleep(600 * time.Millisecond)
	tickPage(t, p)
	row := v.Catalog[30]
	before := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel)
	if before <= 0 {
		t.Fatalf("%s did not advance before the restart", row.Name)
	}
	clickID(t, p, row.Button)
	if !v.Restart[row.Button] {
		t.Fatalf("%s did not flip to the second keyframe name", row.Name)
	}
	if d.manual[row.SampleKey] != nil {
		// A transition restarts by binding its original base again.
		time.Sleep(600 * time.Millisecond)
		tickPage(t, p)
		return
	}
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	after := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel)
	if after >= before {
		t.Fatalf("%s did not replay: %v then %v", row.Name, before, after)
	}
}
