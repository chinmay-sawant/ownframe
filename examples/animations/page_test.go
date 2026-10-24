package main

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

// findBox locates a hit-test box by element id.
func findBox(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}

	return ownframe.Box{}, false
}

// chipHitBox returns the hit-test box for one chip element.
func chipHitBox(t *testing.T, p *ownframe.Page, id string) ownframe.Box {
	t.Helper()

	box, ok := findBox(p.Boxes(), id)
	if !ok {
		t.Fatalf("chip %s has no hit-test box", id)
	}

	return box
}

// opInBox reports whether an operation paints inside box. A text op carries a
// baseline in Y, so its top-left corner is the point to test; a fill carries
// its top-left corner and is tested by its centre.
func opInBox(d *ownframe.Display, box ownframe.Box, op *ownframe.DisplayOp) bool {
	const (
		boxSlack      = 0.5
		centerDivisor = 2.0
	)

	p := d.PointsPerPixel
	x, y := op.X, op.Y
	if op.Kind != layout.DisplayOpText {
		x, y = op.X+op.W/centerDivisor, op.Y+op.H/centerDivisor
	}

	return x >= box.X*p-boxSlack && x <= (box.X+box.W)*p+boxSlack &&
		y >= box.Y*p-boxSlack && y <= (box.Y+box.H)*p+boxSlack
}

// fillOpInBox returns the first fill operation painted inside box.
func fillOpInBox(d *ownframe.Display, box ownframe.Box) *ownframe.DisplayOp {
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind == layout.DisplayOpFillRect && opInBox(d, box, op) {
			return op
		}
	}

	return nil
}

// textOpInBox returns the first text operation painted inside box.
func textOpInBox(d *ownframe.Display, box ownframe.Box) *ownframe.DisplayOp {
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind == layout.DisplayOpText && opInBox(d, box, op) {
			return op
		}
	}

	return nil
}

// xformX is the operation's horizontal translation after its baked matrix.
// A transform-list keyframe writes translate through SetXform and leaves X at
// the laid-out position, so the matrix is the only place the chip moves.
func xformX(op *ownframe.DisplayOp) float64 {
	return op.Transform().E
}

// TestChipGeometry pins the .chip rule that the stylesheet's truncated rule
// used to drop. Only X is pinned; Y depends on how the text above the stage
// wraps, so it is only checked to be laid out below the top.
func TestChipGeometry(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	box := chipHitBox(t, app.Page(), "chip-q1-0")

	const wantW, wantH = 76.0, 32.0
	if box.W != wantW || box.H != wantH {
		t.Fatalf("chip-q1-0 size = %vx%v, want %vx%v", box.W, box.H, wantW, wantH)
	}

	// .chip is absolutely positioned at left:6px inside the track. The track's
	// padding edge sits at x=65: body padding 24 + card border 1 + card padding
	// 22 + stage border 1 + stage padding 16 + track border 1.
	const wantX = 71.0
	if box.X != wantX {
		t.Fatalf("chip-q1-0 X = %v, want %v", box.X, wantX)
	}
	if box.Y <= 0 {
		t.Fatalf("chip-q1-0 Y = %v, want a laid-out position below the top", box.Y)
	}
}

// TestChipLabelKeepsItsColour pins the phase 03 target rule: a background-color
// keyframe writes the chip's fill operation only, never its label text.
// plans/v0.0.3/animations-and-transitions/animation/phase-03-paint-property-coverage.md
func TestChipLabelKeepsItsColour(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	p := app.Page()
	box := chipHitBox(t, p, "chip-q1-0")
	label := textOpInBox(p.Probe().Display(), box)
	if label == nil {
		t.Fatal("expected a text op inside chip-q1-0")
	}
	wantR, wantG, wantB := label.R, label.G, label.B

	if err := p.Testing().Tick(context.Background()); err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	label = textOpInBox(p.Probe().Display(), box)
	if label == nil {
		t.Fatal("the label text op disappeared after a tick")
	}
	if label.R != wantR || label.G != wantG || label.B != wantB {
		t.Fatalf("the background keyframe recoloured the label to %v,%v,%v (was %v,%v,%v)",
			label.R, label.G, label.B, wantR, wantG, wantB)
	}

	fill := fillOpInBox(p.Probe().Display(), box)
	if fill == nil {
		t.Fatal("expected a fill op inside chip-q1-0")
	}
	if fill.R == label.R && fill.G == label.G && fill.B == label.B {
		t.Fatalf("the chip fill and its label share colour %v,%v,%v; the keyframe did not target the fill",
			fill.R, fill.G, fill.B)
	}
}

func TestNewApp(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	if app.Page() == nil {
		t.Fatal("expected non-nil page")
	}

	w, h := app.Page().Size()
	if w != defaultWidth || h != defaultHeight {
		t.Fatalf("expected size %dx%d, got %dx%d", defaultWidth, defaultHeight, w, h)
	}

	if app.view.Total != maxQuestions {
		t.Fatalf("expected total %d, got %d", maxQuestions, app.view.Total)
	}
	if app.view.Score != 0 {
		t.Fatalf("expected initial score 0, got %d", app.view.Score)
	}
	if !app.view.IsPlaying {
		t.Fatal("expected initial play state to be true")
	}

	tracks := 0
	for _, q := range app.view.Questions {
		tracks += len(q.Tracks)
	}
	if tracks == 0 {
		t.Fatal("expected the quiz data to carry animation tracks")
	}
}

// TestAutomaticMotionBinding checks that the page's own publish path bound the
// declared CSS animations, with no application driver involved.
func TestAutomaticMotionBinding(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	if d := app.Page().Probe().Display(); d == nil {
		t.Fatal("expected a non-nil display")
	}

	stats := app.Page().Probe().Motion()
	if stats.Bound == 0 {
		t.Fatal("expected the automatic runtime to bind the chip animations")
	}
	if stats.LastError != "" {
		t.Fatalf("motion bind failed: %s", stats.LastError)
	}
}

// TestAutomaticMotionDrivesPage is the phase 01 acceptance proof: the example
// registers no frame callback, yet two Tick calls sample the bound animations
// and report damage for the next window draw.
func TestAutomaticMotionDrivesPage(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	p := app.Page()
	if p.Probe().Ticking() {
		t.Fatal("the example registered a tick callback")
	}

	d := p.Probe().Display()
	if d == nil {
		t.Fatal("expected a non-nil display")
	}
	chip := fillOpInBox(d, chipHitBox(t, p, "chip-q1-0"))
	if chip == nil {
		t.Fatal("expected a fill op inside chip-q1-0")
	}
	base := xformX(chip)

	ctx := context.Background()
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}
	if !p.Probe().FrameDirty() {
		t.Fatal("the first Tick did not report animation damage")
	}
	if stats := p.Probe().Motion(); stats.Bound == 0 || stats.Sampled == 0 {
		t.Fatalf("first Tick: bound=%d sampled=%d, want both non-zero", stats.Bound, stats.Sampled)
	}
	if !chip.XformSet {
		t.Fatal("the first Tick left the chip without a transform matrix")
	}

	first := xformX(chip)
	if first <= base {
		t.Fatalf("first Tick left the chip translation at %v, want it past the base %v", first, base)
	}
	time.Sleep(5 * time.Millisecond)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second tick failed: %v", err)
	}
	if !p.Probe().FrameDirty() {
		t.Fatal("the second Tick did not report animation damage")
	}
	if second := xformX(chip); second <= first {
		t.Fatalf("chip translation stayed at %v across two Tick calls (base %v)", first, base)
	}
}

func TestHandleOptionClick(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	// Question 1 correct answer is index 1 ("opt-q1-1", steps)
	err = app.handleClick(ctx, ownframe.Box{ID: "opt-q1-1"})
	if err != nil {
		t.Fatalf("click option failed: %v", err)
	}

	if app.view.Score != 1 {
		t.Fatalf("expected score 1 after correct answer, got %d", app.view.Score)
	}
	if app.view.AnsweredCount != 1 {
		t.Fatalf("expected answered count 1, got %d", app.view.AnsweredCount)
	}
	if !app.view.Questions[0].IsCorrect {
		t.Fatal("expected Question 1 to be marked correct")
	}

	// Repeated clicks on answered question should be ignored
	err = app.handleClick(ctx, ownframe.Box{ID: "opt-q1-0"})
	if err != nil {
		t.Fatalf("subsequent click failed: %v", err)
	}
	if app.view.Score != 1 {
		t.Fatalf("expected score to remain 1, got %d", app.view.Score)
	}

	// Question 2 wrong answer is index 0 ("opt-q2-0", correct is index 2)
	err = app.handleClick(ctx, ownframe.Box{ID: "opt-q2-0"})
	if err != nil {
		t.Fatalf("click wrong option failed: %v", err)
	}
	if app.view.Score != 1 {
		t.Fatalf("expected score to remain 1 after wrong answer, got %d", app.view.Score)
	}
	if app.view.AnsweredCount != 2 {
		t.Fatalf("expected answered count 2, got %d", app.view.AnsweredCount)
	}
	if app.view.Questions[1].IsCorrect {
		t.Fatal("expected Question 2 to be marked incorrect")
	}
}

func TestHandlePlayClick(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	// Click play on track 1 of question 1
	err = app.handleClick(ctx, ownframe.Box{ID: "play-q1-1"})
	if err != nil {
		t.Fatalf("click play button failed: %v", err)
	}

	if !app.view.Questions[0].Tracks[1].IsActive {
		t.Fatal("expected Track 1 of Question 1 to be active")
	}
	if app.view.Questions[0].Tracks[0].IsActive {
		t.Fatal("expected Track 0 of Question 1 to be inactive")
	}
}

func TestTogglePlayback(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	if !app.view.IsPlaying {
		t.Fatal("expected playback to start running")
	}

	// Toggle pause
	err = app.handleClick(ctx, ownframe.Box{ID: "btn-toggle-play"})
	if err != nil {
		t.Fatalf("click toggle play failed: %v", err)
	}
	if app.view.IsPlaying {
		t.Fatal("expected playback to be paused")
	}

	// Toggle resume
	err = app.handleClick(ctx, ownframe.Box{ID: "btn-toggle-play"})
	if err != nil {
		t.Fatalf("click toggle play failed: %v", err)
	}
	if !app.view.IsPlaying {
		t.Fatal("expected playback to be resumed")
	}
}

// TestPausedPageDoesNotAdvance proves the automatic runtime reads the play state
// the stylesheet declares. The pause button only toggles the view; the body's
// is-paused class reaches the chips through the existing CSS rule. The redraw
// that posts the class must keep the sampled pose (no snap to the base paint),
// and the ticks after it must leave that pose alone.
func TestPausedPageDoesNotAdvance(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	p := app.Page()
	ctx := context.Background()

	// Give the chip a phase worth freezing: a tick right after the initial
	// redraw can still be at 0%.
	time.Sleep(5 * time.Millisecond)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("tick failed: %v", err)
	}
	box := chipHitBox(t, p, "chip-q1-0")
	chip := fillOpInBox(p.Probe().Display(), box)
	if chip == nil {
		t.Fatal("expected a fill op inside chip-q1-0")
	}
	sampled := xformX(chip)

	if err := app.handleClick(ctx, ownframe.Box{ID: "btn-toggle-play"}); err != nil {
		t.Fatalf("click toggle play failed: %v", err)
	}

	box = chipHitBox(t, p, "chip-q1-0")
	if len(box.Animations) == 0 || box.Animations[0].PlayState != "paused" {
		t.Fatalf("chip-q1-0 animations = %+v, want a paused sample", box.Animations)
	}

	chip = fillOpInBox(p.Probe().Display(), box)
	if chip == nil {
		t.Fatal("expected a fill op inside chip-q1-0 after the pause redraw")
	}
	if got := xformX(chip); got != sampled {
		t.Fatalf("pausing the chip snapped it from translation %v to %v", sampled, got)
	}

	time.Sleep(5 * time.Millisecond)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("tick failed: %v", err)
	}
	if got := xformX(chip); got != sampled {
		t.Fatalf("paused chip moved from translation %v to %v", sampled, got)
	}
	if got := p.Probe().Motion().Active; got != 0 {
		t.Fatalf("paused page reports %d active players, want 0", got)
	}

	// Resuming continues from the frozen phase instead of restarting at 0.
	if err := app.handleClick(ctx, ownframe.Box{ID: "btn-toggle-play"}); err != nil {
		t.Fatalf("click toggle play failed: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("tick after resume failed: %v", err)
	}
	box = chipHitBox(t, p, "chip-q1-0")
	chip = fillOpInBox(p.Probe().Display(), box)
	if chip == nil {
		t.Fatal("expected a fill op inside chip-q1-0 after the resume redraw")
	}
	if got := p.Probe().Motion().Active; got == 0 {
		t.Fatal("the resumed page reports no active players")
	}
	if got := xformX(chip); got <= sampled {
		t.Fatalf("resumed chip stayed at translation %v (paused at %v), want it to continue", got, sampled)
	}
}

func TestResetQuiz(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	// Answer question 1
	_ = app.handleClick(ctx, ownframe.Box{ID: "opt-q1-1"})
	if app.view.Score != 1 || app.view.AnsweredCount != 1 {
		t.Fatalf("expected answered question before reset")
	}

	// Click reset
	err = app.handleClick(ctx, ownframe.Box{ID: "btn-reset"})
	if err != nil {
		t.Fatalf("click reset failed: %v", err)
	}

	if app.view.Score != 0 {
		t.Fatalf("expected score 0 after reset, got %d", app.view.Score)
	}
	if app.view.AnsweredCount != 0 {
		t.Fatalf("expected answered count 0 after reset, got %d", app.view.AnsweredCount)
	}
	for i, q := range app.view.Questions {
		if q.Answered || q.SelectedIndex != -1 {
			t.Fatalf("expected question %d to be unselected after reset", i)
		}
	}
}

func TestCategoryFilter(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	if app.view.ActiveFilter != "all" {
		t.Fatalf("expected default filter 'all', got %q", app.view.ActiveFilter)
	}

	err = app.handleClick(ctx, ownframe.Box{ID: "filter-timing"})
	if err != nil {
		t.Fatalf("click filter timing failed: %v", err)
	}
	if app.view.ActiveFilter != "timing" {
		t.Fatalf("expected active filter 'timing', got %q", app.view.ActiveFilter)
	}

	err = app.handleClick(ctx, ownframe.Box{ID: "filter-all"})
	if err != nil {
		t.Fatalf("click filter all failed: %v", err)
	}
	if app.view.ActiveFilter != "all" {
		t.Fatalf("expected active filter 'all', got %q", app.view.ActiveFilter)
	}
}
