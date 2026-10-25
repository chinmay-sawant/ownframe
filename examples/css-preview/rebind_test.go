package main

import (
	"testing"
	"time"
)

// TestSecondAnimationRebindsFirstAcrossRedraw checks that a redraw caused by
// another Start rebuilds the first card's player at the clock it reached, not
// at zero.
func TestSecondAnimationRebindsFirstAcrossRedraw(t *testing.T) {
	p, _, v := fixturePage(t)
	first := catalogRowByName(t, v, "animation-name")
	clickID(t, p, first.Button)
	time.Sleep(10 * time.Millisecond)
	tickPage(t, p)
	before := xformX(targetOp(t, p, first.TargetID), p.Probe().Display().PointsPerPixel)
	if before <= 0 {
		t.Fatalf("the first sample did not advance: %v", before)
	}
	rebinds := p.Probe().Motion().Rebinds
	second := catalogRowByName(t, v, "-webkit-animation")
	clickID(t, p, second.Button)
	if p.Probe().Motion().Rebinds <= rebinds {
		t.Fatal("the second click did not rebind the first player")
	}
	time.Sleep(5 * time.Millisecond)
	tickPage(t, p)
	after := xformX(targetOp(t, p, first.TargetID), p.Probe().Display().PointsPerPixel)
	if after < before {
		t.Fatalf("the redraw restarted the first sample: %v then %v", before, after)
	}
}
