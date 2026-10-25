package main

import (
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe"
)

// costWindowFrames is how many frames one window run measures.
const costWindowFrames = 90

// TestCostWindowFrames runs the real window over the preview page and
// reports what one frame costs. OWNFRAME_COST_MODE picks the case:
// anim-full renders all 43 rows and starts all 43 previews, anim plays
// every preview the way Play All does now, idle keeps the frame loop
// alive with nothing running. The window runs once per process, so the
// cases are three runs.
func TestCostWindowFrames(t *testing.T) {
	if os.Getenv("OWNFRAME_COST_WINDOW") != "1" {
		t.Skip("needs a display; set OWNFRAME_COST_WINDOW=1")
	}
	switch os.Getenv("OWNFRAME_COST_MODE") {
	case "anim":
		costWindowRun(t, "anim", 1)
	case "anim-full":
		costWindowRun(t, "anim-full", 40)
	case "idle":
		costWindowRun(t, "idle", 0)
	default:
		t.Skip("set OWNFRAME_COST_MODE=anim, anim-full, or idle")
	}
}

// costWindowRun opens the real window and measures its frames. The window
// is the shipping one; the harness only wraps it to time the draw calls.
func costWindowRun(tb testing.TB, mode string, target int) {
	tb.Helper()
	ctx, p, v, d, _ := windowBenchPageSize(tb, 1908, 999)
	switch mode {
	case "anim":
		costPlayAll(ctx, p, tb)
	case "anim-full":
		costFullCatalog(ctx, p, v, d, tb)
		costPlayAll(ctx, p, tb)
	}
	g := &costGame{inner: ownframe.NewGame(ctx, p)}
	c := &costCollector{p: p, d: d, v: v, g: g, warm: 3}
	p.SetTick(c.tick)
	ebiten.SetWindowSize(1908, 999)
	ebiten.SetWindowTitle("ownframe cost")
	if err := ebiten.RunGame(g); err != nil {
		tb.Fatal(err)
	}
	costReportFrames(tb, mode, c.costPick(target))
	tb.Logf("%s: rows=%d ops=%d boxes=%d draws=%d", mode, len(v.Rows),
		p.Probe().Stats().Ops, p.Probe().Stats().Boxes, g.draws)
	motion := p.Probe().Motion()
	st := p.Probe().Stats()
	tb.Logf("%s motion: bound=%d active=%d sampled=%d rebinds=%d restamped=%d gaps=%d "+
		"transitions started=%d active=%d sampled=%d writes=%d dirty=%d",
		mode, motion.Bound, motion.Active, motion.Sampled, motion.Rebinds, motion.Restamped,
		len(motion.Gaps), motion.TransitionsStarted, motion.TransitionsActive,
		motion.TransitionsSampled, st.ChangedOps, st.DirtyOps)
	tb.Logf("%s actives: %v", mode, costActives(c.frames))
	tb.Logf("%s draws: %v", mode, costDraws(c.frames))
}
