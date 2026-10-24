package main

import (
	"strings"
	"testing"
	"time"
)

// TestSecondStartRestartsAManualTransition checks that a transition card
// replays from its captured base when Start is pressed again.
func TestSecondStartRestartsAManualTransition(t *testing.T) {
	p, d, v := fixturePage(t)
	row := catalogRowByName(t, v, "transition")
	clickID(t, p, row.Button)
	time.Sleep(120 * time.Millisecond)
	tickPage(t, p)
	before := targetPaints(t, p, row.TargetID)
	clickID(t, p, row.Button)
	time.Sleep(5 * time.Millisecond)
	tickPage(t, p)
	if !v.Restart[row.Button] {
		t.Fatal("the second Start did not flip the restart state")
	}
	if v.Restarting[row.Button] {
		t.Fatal("the restart hide class survived the commit")
	}
	if row := catalogRowByName(t, v, "transition"); !strings.Contains(row.Class, "restart-b") {
		t.Fatalf("card class = %q, want the second keyframe variant", row.Class)
	}
	if s := d.manual[row.SampleKey]; s == nil || !s.active {
		t.Fatal("the transition did not replay")
	}
	after := targetPaints(t, p, row.TargetID)
	if samePaint(before, after) {
		t.Fatal("the transition stayed at its previous paint")
	}
}

// TestSecondStartRestartsATriggerCard checks the manual trigger path across
// a restart. The hide publish gives the new keyframe name a fresh clock.
func TestSecondStartRestartsATriggerCard(t *testing.T) {
	p, _, v := fixturePage(t)
	row := catalogRowByName(t, v, "animation-trigger")
	clickID(t, p, row.Button)
	time.Sleep(120 * time.Millisecond)
	tickPage(t, p)
	before := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel)
	clickID(t, p, row.Button)
	time.Sleep(5 * time.Millisecond)
	tickPage(t, p)
	after := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel)
	if after >= before {
		t.Fatalf("the trigger card did not replay: %v then %v", before, after)
	}
}
