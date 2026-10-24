package main

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const (
	settleDelay = 10 * time.Millisecond
	secondDelay = 15 * time.Millisecond
)

// showcaseTargetIDs lists the animated element of every card in the gallery.
// The list is the test's definition of "all cards": a target that is not here
// is not sampled, and an entry that is not in animations.html fails the hit-box
// lookup.
func showcaseTargetIDs() []string {
	return []string{
		"show-ease", "show-linear", "show-bezier", "show-steps",
		"show-direction-normal", "show-direction-reverse",
		"show-direction-alternate", "show-direction-alternate-reverse",
		"show-delay-positive", "show-delay-negative",
		"show-fill-none", "show-fill-forwards", "show-fill-both",
		"show-iteration-once", "show-iteration-three", "show-iteration-infinite",
		"show-translate", "show-rotate", "show-scale", "show-skew", "show-matrix",
		"show-rotate-x", "show-rotate-y", "show-translate-z", "show-matrix3d",
		"show-perspective", "show-perspective-origin",
		"show-opacity", "show-background", "show-color", "show-border",
		"show-corner", "show-webkit",
	}
}

// TestEveryCardExplainsItsEffect keeps the gallery a teaching page: every card
// must say what the viewer should see, not just name the property.
func TestEveryCardExplainsItsEffect(t *testing.T) {
	want := len(showcaseTargetIDs())
	if got := strings.Count(animationsHTML, `class="expect"`); got != want {
		t.Fatalf("expectation blocks = %d, want %d", got, want)
	}
	if got := strings.Count(animationsHTML, `class="expect-label"`); got != want {
		t.Fatalf("expectation labels = %d, want %d", got, want)
	}
	if strings.Contains(animationsHTML, `class="expect"><span class="expect-label">Expect</span> </p>`) {
		t.Fatal("an expectation block is empty")
	}
}

// newTestPage builds the gallery and requires a replayable display.
func newTestPage(t *testing.T) *ownframe.Page {
	t.Helper()

	app, err := newApp(false)
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	p := app.Page()
	if p.Probe().Display() == nil {
		t.Fatal("expected a display list, got a bitmap fallback")
	}

	return p
}

// boxByID returns the hit-test box for one element id.
func boxByID(t *testing.T, p *ownframe.Page, id string) ownframe.Box {
	t.Helper()

	for _, box := range p.Boxes() {
		if box.ID == id {
			return box
		}
	}

	t.Fatalf("element %q has no hit-test box", id)

	return ownframe.Box{}
}

// opsInBox returns the operations the box's subtree owns, in display order.
// A box carries its subtree's operation IDs, so a card's fill and its label
// text can be read separately without guessing from geometry.
func opsInBox(d *ownframe.Display, box ownframe.Box) []*ownframe.DisplayOp {
	owned := make(map[uint64]struct{}, len(box.OperationIDs))
	for _, id := range box.OperationIDs {
		owned[id] = struct{}{}
	}

	ops := make([]*ownframe.DisplayOp, 0, len(box.OperationIDs))

	for i := range d.Ops {
		if _, ok := owned[d.Ops[i].ID]; ok {
			ops = append(ops, &d.Ops[i])
		}
	}

	return ops
}

// firstFill returns the first background fill among ops.
func firstFill(ops []*ownframe.DisplayOp) *ownframe.DisplayOp {
	for _, op := range ops {
		if op.Kind == ownframe.DisplayOpFillRect {
			return op
		}
	}

	return nil
}

// firstText returns the first shaped text run among ops.
func firstText(ops []*ownframe.DisplayOp) *ownframe.DisplayOp {
	for _, op := range ops {
		if op.Kind == ownframe.DisplayOpText {
			return op
		}
	}

	return nil
}

// requireFill returns the first fill op owned by the element id.
func requireFill(t *testing.T, p *ownframe.Page, id string) *ownframe.DisplayOp {
	t.Helper()

	fill := firstFill(opsInBox(p.Probe().Display(), boxByID(t, p, id)))
	if fill == nil {
		t.Fatalf("%s has no fill operation", id)
	}

	return fill
}

// requireText returns the first text op owned by the element id.
func requireText(t *testing.T, p *ownframe.Page, id string) *ownframe.DisplayOp {
	t.Helper()

	text := firstText(opsInBox(p.Probe().Display(), boxByID(t, p, id)))
	if text == nil {
		t.Fatalf("%s has no text operation", id)
	}

	return text
}

// colourOf copies an operation's sRGB channels.
func colourOf(op *ownframe.DisplayOp) [3]float64 {
	return [3]float64{op.R, op.G, op.B}
}

// clickID clicks the centre of an element's hit box, the same hit test the
// window runs for a pointer click.
func clickID(t *testing.T, p *ownframe.Page, id string) {
	t.Helper()

	box := boxByID(t, p, id)
	if err := p.Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatalf("click %s: %v", id, err)
	}
}

// translationX returns the operation's baked horizontal translation. A 2D
// transform keyframe writes SetXform and leaves X at the laid-out position.
func translationX(t *testing.T, p *ownframe.Page, id string) float64 {
	t.Helper()

	return requireFill(t, p, id).Transform().E
}

// pathHalves returns the background fill path and the border stroke path of a
// corner-shape element. Both are one operation kind, so the Stroke flag is what
// separates them.
func pathHalves(t *testing.T, p *ownframe.Page, id string) (*ownframe.DisplayOp, *ownframe.DisplayOp) {
	t.Helper()

	var fill, stroke *ownframe.DisplayOp

	for _, op := range opsInBox(p.Probe().Display(), boxByID(t, p, id)) {
		corner, ok := op.CornerPathValue()
		if !ok || corner == nil {
			continue
		}

		if corner.Stroke {
			stroke = op
		} else {
			fill = op
		}
	}

	if fill == nil || stroke == nil {
		t.Fatalf("%s path halves: fill=%v stroke=%v, want both", id, fill != nil, stroke != nil)
	}

	return fill, stroke
}

// projection is the painted form of one operation. A re-projected 3D op may
// change its matrix, its geometry, or its final display path, so the snapshot
// copies all three. The display list owns those arrays, so a comparison against
// an earlier frame needs a copy.
type projection struct {
	x, y, w, h float64
	xformSet   bool
	xform      [6]float64
	points     [][2]float64
}

func projectionOf(t *testing.T, p *ownframe.Page, id string) projection {
	t.Helper()

	ops := opsInBox(p.Probe().Display(), boxByID(t, p, id))
	if len(ops) == 0 {
		t.Fatalf("%s owns no operation", id)
	}

	op := ops[0]
	matrix := op.Transform()
	out := projection{
		x:        op.X,
		y:        op.Y,
		w:        op.W,
		h:        op.H,
		xformSet: op.XformSet,
		xform:    [6]float64{matrix.A, matrix.B, matrix.C, matrix.D, matrix.E, matrix.F},
	}

	if path, ok := op.PathValue(); ok {
		out.points = make([][2]float64, len(path.Points))
		for i, point := range path.Points {
			out.points[i] = [2]float64{point.X, point.Y}
		}
	}

	return out
}

func TestNewApp(t *testing.T) {
	app, err := newApp(false)
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

	if !app.view.IsPlaying {
		t.Fatal("expected the gallery to start playing")
	}
}

// TestEveryCardBindsWithoutATickCallback is the gallery's contract: every card
// declares its animation in CSS, the automatic runtime binds it at the initial
// redraw, and no Go callback drives it.
func TestEveryCardBindsWithoutATickCallback(t *testing.T) {
	p := newTestPage(t)

	if p.Probe().Ticking() {
		t.Fatal("the gallery registered a tick callback")
	}

	stats := p.Probe().Motion()
	if stats.Bound == 0 {
		t.Fatal("the automatic runtime bound no players")
	}
	if stats.LastError != "" {
		t.Fatalf("motion bind failed: %s", stats.LastError)
	}
	if len(stats.Gaps) > 0 {
		t.Fatalf("cards declared properties the driver does not place: %+v", stats.Gaps)
	}
	if !stats.NeedsRetainedState {
		t.Fatal("the 3D and perspective cards did not ask for retained layout state")
	}

	want := len(showcaseTargetIDs())
	if got := strings.Count(animationsHTML, `id="show-`); got != want {
		t.Fatalf("the gallery has %d showcase targets, the test lists %d", got, want)
	}

	for _, id := range showcaseTargetIDs() {
		if box := boxByID(t, p, id); len(box.Animations) == 0 {
			t.Fatalf("%s carries no declared animation", id)
		}
	}
}

// TestTranslateCardAdvances proves the 2D card moves through its operation
// matrix across two ticks with no application callback.
func TestTranslateCardAdvances(t *testing.T) {
	p := newTestPage(t)
	ctx := context.Background()

	base := translationX(t, p, "show-translate")

	time.Sleep(settleDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}
	if !p.Probe().FrameDirty() {
		t.Fatal("the first tick did not report animation damage")
	}

	op := requireFill(t, p, "show-translate")
	if !op.XformSet {
		t.Fatal("the translate card carries no baked matrix")
	}

	first := translationX(t, p, "show-translate")
	if first <= base {
		t.Fatalf("first tick left the translation at %v, want it past the base %v", first, base)
	}

	time.Sleep(secondDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second tick failed: %v", err)
	}

	if second := translationX(t, p, "show-translate"); second <= first {
		t.Fatalf("translation stayed at %v across the second tick (first %v)", second, first)
	}
}

// TestThreeDCardRestampsAndMoves samples the rotateY card through two ticks.
// Restamped is the applied re-projection request, so it survives the rebind
// that adopts the re-projected display.
func TestThreeDCardRestampsAndMoves(t *testing.T) {
	p := newTestPage(t)
	ctx := context.Background()

	base := projectionOf(t, p, "show-rotate-y")

	time.Sleep(settleDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}

	firstStats := p.Probe().Motion()
	if firstStats.Restamped == 0 {
		t.Fatalf("the first tick did not restamp the 3D card: needsRestamp=%d restampError=%q",
			firstStats.NeedsRestamp, firstStats.RestampError)
	}
	if firstStats.RestampError != "" {
		t.Fatalf("first restamp failed: %s", firstStats.RestampError)
	}

	first := projectionOf(t, p, "show-rotate-y")
	if reflect.DeepEqual(first, base) {
		t.Fatalf("the first tick kept the base projection %+v", base)
	}

	time.Sleep(secondDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second tick failed: %v", err)
	}

	stats := p.Probe().Motion()
	if stats.Restamped == 0 {
		t.Fatalf("the second tick did not restamp the 3D card: needsRestamp=%d restampError=%q",
			stats.NeedsRestamp, stats.RestampError)
	}
	if stats.RestampError != "" {
		t.Fatalf("second restamp failed: %s", stats.RestampError)
	}

	if second := projectionOf(t, p, "show-rotate-y"); reflect.DeepEqual(second, first) {
		t.Fatalf("the second tick kept the first projection %+v", first)
	}
}

// TestBackgroundCardLeavesItsLabelAlone checks the nearest-owner rule: a
// background-colour keyframe writes the chip's fill, never its text run.
func TestBackgroundCardLeavesItsLabelAlone(t *testing.T) {
	p := newTestPage(t)
	ctx := context.Background()

	baseFill := colourOf(requireFill(t, p, "show-background"))
	baseInk := colourOf(requireText(t, p, "show-background"))

	time.Sleep(settleDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}

	firstFill := colourOf(requireFill(t, p, "show-background"))
	firstInk := colourOf(requireText(t, p, "show-background"))
	if firstFill == baseFill {
		t.Fatalf("the background keyframe did not recolour the fill, still %v", baseFill)
	}
	if firstInk != baseInk {
		t.Fatalf("the background keyframe recoloured the label: base %v, first %v", baseInk, firstInk)
	}

	time.Sleep(secondDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second tick failed: %v", err)
	}

	secondFill := colourOf(requireFill(t, p, "show-background"))
	secondInk := colourOf(requireText(t, p, "show-background"))
	if secondFill == baseFill {
		t.Fatalf("the background keyframe left the fill at %v", baseFill)
	}
	if secondInk != baseInk {
		t.Fatalf("the background keyframe recoloured the label: base %v, second %v", baseInk, secondInk)
	}
	if secondFill == secondInk {
		t.Fatalf("the fill and its label share colour %v; the keyframe did not target the fill alone", secondFill)
	}
}

// TestCornerPathCardRecoloursTheFill pins the path half of the paint coverage:
// a background keyframe writes the corner path fill and leaves its border
// stroke and command list intact.
func TestCornerPathCardRecoloursTheFill(t *testing.T) {
	p := newTestPage(t)
	ctx := context.Background()

	baseFill, baseStroke := pathHalves(t, p, "show-corner")
	baseFillColour := colourOf(baseFill)
	baseStrokeColour := colourOf(baseStroke)

	time.Sleep(settleDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}

	firstFill, firstStroke := pathHalves(t, p, "show-corner")
	if colourOf(firstFill) == baseFillColour {
		t.Fatalf("the path fill stayed at %v", baseFillColour)
	}
	if colourOf(firstStroke) != baseStrokeColour {
		t.Fatalf("the path border moved: base %v, first %v",
			baseStrokeColour, colourOf(firstStroke))
	}

	time.Sleep(secondDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second tick failed: %v", err)
	}

	secondFill, secondStroke := pathHalves(t, p, "show-corner")
	if colourOf(secondFill) == baseFillColour {
		t.Fatalf("the path fill stayed at %v after two ticks", baseFillColour)
	}
	if colourOf(secondStroke) != baseStrokeColour {
		t.Fatalf("the path border moved: base %v, second %v",
			baseStrokeColour, colourOf(secondStroke))
	}

	for _, op := range []*ownframe.DisplayOp{secondFill, secondStroke} {
		corner, ok := op.CornerPathValue()
		if !ok || corner == nil || len(corner.Cmds) == 0 {
			t.Fatal("a colour sample dropped the path command list")
		}
	}
}

// TestPauseKeepsThePoseAndResumeContinues toggles the one control. The click
// redraw posts the body class with the declared paint; the next tick samples
// the carried clock, the same tick the window runs before it draws. From then
// on the paused ticks leave the pose alone, and resuming continues from it
// instead of restarting at zero.
func TestPauseKeepsThePoseAndResumeContinues(t *testing.T) {
	app, err := newApp(false)
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	p := app.Page()
	ctx := context.Background()

	time.Sleep(settleDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	sampled := translationX(t, p, "show-translate")

	clickID(t, p, togglePlayID)
	if app.view.IsPlaying {
		t.Fatal("the pause button left the gallery playing")
	}

	box := boxByID(t, p, "show-translate")
	if len(box.Animations) == 0 || box.Animations[0].PlayState != "paused" {
		t.Fatalf("show-translate animations = %+v, want a paused sample", box.Animations)
	}

	time.Sleep(secondDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("tick while paused failed: %v", err)
	}

	paused := translationX(t, p, "show-translate")
	if paused < sampled {
		t.Fatalf("pausing moved the card back from %v to %v", sampled, paused)
	}

	time.Sleep(secondDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second tick while paused failed: %v", err)
	}
	if got := translationX(t, p, "show-translate"); got != paused {
		t.Fatalf("paused card moved from translation %v to %v", paused, got)
	}

	clickID(t, p, togglePlayID)
	if !app.view.IsPlaying {
		t.Fatal("the play button left the gallery paused")
	}

	time.Sleep(secondDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("tick after resume failed: %v", err)
	}
	if got := translationX(t, p, "show-translate"); got <= paused {
		t.Fatalf("resumed card stayed at translation %v (paused at %v), want it to continue", got, paused)
	}
}
