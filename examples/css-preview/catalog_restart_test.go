package main

import (
	"math"
	"strings"
	"testing"
	"time"
)

// transitionRestartDuration is the transition card's declared
// transition-duration in testdata/css-preview/styles.css.
const transitionRestartDuration = 900 * time.Millisecond

// TestSecondStartRestartsAManualTransition checks that a transition card
// replays from its captured base when Start is pressed again.
//
// As with the trigger card, the restart is checked at the clock. The second
// Start publishes twice and prepareManual stamps the manual clock before both,
// so a tick after that click can land past the end of the 900 ms transition on
// a loaded machine. Liveness after such a tick is not the contract. The
// contract is the rebind: a new renderer takes over at a later origin, that
// origin shows the captured base, the midpoint shows a paint between the
// endpoints, and the full duration shows the captured destination.
func TestSecondStartRestartsAManualTransition(t *testing.T) {
	p, d, v := fixturePage(t)
	row := catalogRowByName(t, v, "transition")
	clickID(t, p, row.Button)
	first := d.manual[row.SampleKey]
	if first == nil || !first.active || first.renderer == nil {
		t.Fatal("the first Start did not bind a transition renderer")
	}
	// The sample struct is reused across the restart, so snapshot what the
	// second click must replace before it mutates the fields.
	firstRenderer, firstOrigin := first.renderer, first.origin
	time.Sleep(120 * time.Millisecond)
	tickPage(t, p)
	before := targetPaints(t, p, row.TargetID)

	clickID(t, p, row.Button)
	if !v.Restart[row.Button] {
		t.Fatal("the second Start did not flip the restart state")
	}
	if v.Restarting[row.Button] {
		t.Fatal("the restart hide class survived the commit")
	}
	if row := catalogRowByName(t, v, "transition"); !strings.Contains(row.Class, "restart-b") {
		t.Fatalf("card class = %q, want the second keyframe variant", row.Class)
	}
	restarted := d.manual[row.SampleKey]
	if restarted == nil || !restarted.active || restarted.renderer == nil {
		t.Fatal("the transition did not replay")
	}
	if restarted.renderer == firstRenderer {
		t.Fatal("the restart kept the renderer the first click bound")
	}
	if !restarted.origin.After(firstOrigin) {
		t.Fatalf("restart origin %v is not after the first origin %v", restarted.origin, firstOrigin)
	}
	if samePaint(before, restarted.from) {
		t.Fatalf("the first sample is still the captured base %v", before)
	}

	// Step the new renderer on its own clock instead of waiting for the page
	// tick, so the phase does not depend on how long the two publishes take.
	if !restarted.renderer.Step(restarted.origin) {
		t.Fatal("the restarted transition was already over at its origin")
	}
	if base := targetPaints(t, p, row.TargetID); !samePaint(base, restarted.from) {
		t.Fatalf("the restarted transition started at %v, want the captured base %v", base, restarted.from)
	}
	if !restarted.renderer.Step(restarted.origin.Add(transitionRestartDuration / 2)) {
		t.Fatal("the restarted transition stopped before its midpoint")
	}
	mid := targetPaints(t, p, row.TargetID)
	if samePaint(mid, restarted.from) || samePaint(mid, restarted.to) {
		t.Fatalf("the midpoint paint %v is an endpoint, want a value between %v and %v",
			mid, restarted.from, restarted.to)
	}
	if restarted.renderer.Step(restarted.origin.Add(transitionRestartDuration)) {
		t.Fatal("the restarted transition is still active after its full duration")
	}
	if end := targetPaints(t, p, row.TargetID); !samePaint(end, restarted.to) {
		t.Fatalf("the transition ended at %v, want the captured destination %v", end, restarted.to)
	}
}

// TestSecondStartRestartsATriggerCard checks the manual trigger path across
// a restart. The hide publish gives the new keyframe name a fresh clock.
//
// The two samples are not ordered by click. The second Start publishes twice
// (the hide pass, then the variant pass), and prepareManual stamps the manual
// clock before both, so on a loaded machine the restarted player has already
// run for longer than the first click's 120 ms lead and legitimately shows a
// value past the first sample. The restart contract is the clock, not a
// smaller number: the republished display is back at the declared base paint,
// the first player is released, and the new player starts at the second
// click's origin with no carried phase or offset.
func TestSecondStartRestartsATriggerCard(t *testing.T) {
	p, d, v := fixturePage(t)
	row := catalogRowByName(t, v, "animation-trigger")
	clickID(t, p, row.Button)
	first := d.manual[row.SampleKey]
	if first == nil || !first.active || first.player == nil {
		t.Fatal("the first Start did not bind a trigger player")
	}
	// The sample struct is reused across the restart, so snapshot what the
	// second click must replace before it mutates the fields.
	firstPlayer, firstOrigin := first.player, first.origin
	time.Sleep(120 * time.Millisecond)
	tickPage(t, p)
	before := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel)
	if before <= 0 {
		t.Fatalf("the first sample = %v, want the card to move off its base", before)
	}

	clickID(t, p, row.Button)
	if !v.Restart[row.Button] {
		t.Fatal("the second Start did not flip the restart state")
	}
	if row := catalogRowByName(t, v, "animation-trigger"); !strings.Contains(row.Class, "restart-b") {
		t.Fatalf("card class = %q, want the second keyframe variant", row.Class)
	}
	// The publish that carries the second click lays the card out from CSS, so
	// the pose the first run reached is gone before the restarted player
	// samples. That is the visible half of the restart and needs no clock.
	if got := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel); math.Abs(got) > .01 {
		t.Fatalf("the restart left the target at %v px, want the declared base at 0", got)
	}
	restarted := d.manual[row.SampleKey]
	if restarted == nil || !restarted.active || restarted.player == nil {
		t.Fatal("the restart did not bind a trigger player")
	}
	if restarted.player == firstPlayer {
		t.Fatal("the restart kept the player the first click bound")
	}
	if !restarted.origin.After(firstOrigin) {
		t.Fatalf("restart origin %v is not after the first origin %v", restarted.origin, firstOrigin)
	}
	clock, ok := restarted.player.Clock()
	if !ok || !clock.Start.Equal(restarted.origin) || clock.Offset != 0 {
		t.Fatalf("restarted clock = %+v (started %v), want a fresh clock at %v", clock, ok, restarted.origin)
	}

	// The tick drives the new player forward from the base instead of leaving
	// it at the pose the first run reached.
	time.Sleep(5 * time.Millisecond)
	tickPage(t, p)
	after := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel)
	if after <= 0 {
		t.Fatalf("the restarted player stayed at its base: %v", after)
	}
}
