package main

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const (
	settleDelay = 10 * time.Millisecond
	secondDelay = 15 * time.Millisecond
	// transitionProbeDelay is short enough to stay inside the two-second
	// transition the cards declare, and long enough that the sampled value sits
	// strictly between the endpoints after the first sample.
	transitionProbeDelay = 400 * time.Millisecond
)

// newTestPage builds the conformance page and requires a replayable display.
func newTestPage(t *testing.T) *ownframe.Page {
	t.Helper()

	p, err := newPage()
	if err != nil {
		t.Fatalf("newPage failed: %v", err)
	}

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

// between reports that value lies strictly between the two endpoints. A
// transition that snapped would equal one of them.
func between(value, a, b float64) bool {
	if a > b {
		a, b = b, a
	}

	return value > a && value < b
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

// TestPageBindsWithoutATickCallback proves the page is driven only by its CSS.
func TestPageBindsWithoutATickCallback(t *testing.T) {
	p := newTestPage(t)

	if p.Probe().Ticking() {
		t.Fatal("the conformance page registered a tick callback")
	}

	stats := p.Probe().Motion()
	if stats.Bound == 0 {
		t.Fatal("the automatic runtime bound no players")
	}
	if stats.LastError != "" {
		t.Fatalf("motion bind failed: %s", stats.LastError)
	}
	if !stats.NeedsRetainedState {
		t.Fatal("the 3D and perspective cards did not ask for retained layout state")
	}

	ids := []string{
		"timing-linear", "timing-ease-in-out", "timing-bezier", "timing-steps",
		"dir-normal", "dir-reverse", "dir-alternate",
		"delay-zero", "delay-positive", "delay-negative",
		"fill-none", "fill-forwards", "fill-both",
		"iteration-one", "iteration-three", "iteration-infinite",
		"translate-tile", "scale-tile", "rotate-tile", "skew-tile", "matrix-tile",
		"rotate-x-tile", "rotate-y-tile", "translate-z-tile", "matrix-3d-stage",
		"perspective-stage", "perspective-origin-stage",
		"opacity-tile",
		"colour-bg-tile", "colour-ink-tile",
		"edge-tile", "path-tile",
		"play-pause-playing", "play-pause-paused",
	}

	for _, id := range ids {
		if box := boxByID(t, p, id); len(box.Animations) == 0 {
			t.Fatalf("%s carries no declared animation", id)
		}
	}
}

// TestTwoTicksSampleEveryShippedClass builds the page, advances two ticks, and
// checks one representative of every class the automatic driver writes: the 2D
// matrix path, the 3D re-projection path, fill-only colour writes, the corner
// path, and the paused clock.
func TestTwoTicksSampleEveryShippedClass(t *testing.T) {
	p := newTestPage(t)
	ctx := context.Background()

	base2D := translationX(t, p, "translate-tile")
	baseProjection := projectionOf(t, p, "rotate-y-tile")
	baseMatrix3D := projectionOf(t, p, "matrix-3d-tile")
	baseFill := colourOf(requireFill(t, p, "colour-bg-tile"))
	baseInk := colourOf(requireText(t, p, "colour-bg-tile"))

	basePathFill, basePathStroke := pathHalves(t, p, "path-tile")
	basePathFillColour := colourOf(basePathFill)
	basePathStrokeColour := colourOf(basePathStroke)

	time.Sleep(settleDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}

	firstStats := p.Probe().Motion()
	first2D := translationX(t, p, "translate-tile")
	firstProjection := projectionOf(t, p, "rotate-y-tile")
	firstMatrix3D := projectionOf(t, p, "matrix-3d-tile")
	firstFill := colourOf(requireFill(t, p, "colour-bg-tile"))
	firstInk := colourOf(requireText(t, p, "colour-bg-tile"))
	firstPathFill, firstPathStroke := pathHalves(t, p, "path-tile")
	firstPathFillColour := colourOf(firstPathFill)
	firstPathStrokeColour := colourOf(firstPathStroke)
	firstPaused := translationX(t, p, "play-pause-paused")
	firstPlaying := translationX(t, p, "play-pause-playing")

	time.Sleep(secondDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second tick failed: %v", err)
	}

	stats := p.Probe().Motion()
	second2D := translationX(t, p, "translate-tile")
	secondProjection := projectionOf(t, p, "rotate-y-tile")
	secondMatrix3D := projectionOf(t, p, "matrix-3d-tile")
	secondFill := colourOf(requireFill(t, p, "colour-bg-tile"))
	secondInk := colourOf(requireText(t, p, "colour-bg-tile"))
	secondPathFill, secondPathStroke := pathHalves(t, p, "path-tile")
	secondPathFillColour := colourOf(secondPathFill)
	secondPathStrokeColour := colourOf(secondPathStroke)
	secondPaused := translationX(t, p, "play-pause-paused")
	secondPlaying := translationX(t, p, "play-pause-playing")

	t.Run("2D card writes a matrix", func(t *testing.T) {
		op := requireFill(t, p, "translate-tile")
		if !op.XformSet {
			t.Fatal("the 2D translate card carries no baked matrix")
		}
		if second2D <= base2D {
			t.Fatalf("2D translation stayed at %v (base %v)", second2D, base2D)
		}
		if second2D <= first2D {
			t.Fatalf("2D translation stayed at %v across the second tick (first %v)", second2D, first2D)
		}
	})

	t.Run("3D card restamps and moves on the second tick", func(t *testing.T) {
		// NeedsRestamp counts the last sample, but the rebind that adopts a
		// re-projected display rebuilds the players and resets that field.
		// Restamped is the applied request, so it is the counter that
		// survives the tick.
		if firstStats.Restamped == 0 {
			t.Fatalf("the first tick did not restamp the 3D card: needsRestamp=%d restampError=%q",
				firstStats.NeedsRestamp, firstStats.RestampError)
		}
		if firstStats.RestampError != "" {
			t.Fatalf("first restamp failed: %s", firstStats.RestampError)
		}
		if reflect.DeepEqual(firstProjection, baseProjection) {
			t.Fatalf("the first tick kept the base projection %+v", baseProjection)
		}

		if stats.Restamped == 0 {
			t.Fatalf("the second tick did not restamp the 3D card: needsRestamp=%d restampError=%q",
				stats.NeedsRestamp, stats.RestampError)
		}
		if stats.RestampError != "" {
			t.Fatalf("second restamp failed: %s", stats.RestampError)
		}
		if reflect.DeepEqual(secondProjection, firstProjection) {
			t.Fatalf("the second tick kept the first projection %+v", firstProjection)
		}
		if reflect.DeepEqual(firstMatrix3D, baseMatrix3D) {
			t.Fatalf("the animated camera kept the matrix3d subtree's base projection %+v", baseMatrix3D)
		}
		if reflect.DeepEqual(secondMatrix3D, firstMatrix3D) {
			t.Fatalf("the second tick kept the matrix3d subtree's first projection %+v", firstMatrix3D)
		}
	})

	t.Run("colour card leaves the label untouched", func(t *testing.T) {
		if firstFill == baseFill {
			t.Fatalf("the background keyframe did not recolour the fill, still %v", baseFill)
		}
		if secondFill == baseFill {
			t.Fatalf("the background keyframe left the fill at %v", baseFill)
		}
		if firstInk != baseInk || secondInk != baseInk {
			t.Fatalf("the background keyframe recoloured the label: base %v, first %v, second %v",
				baseInk, firstInk, secondInk)
		}
		if secondFill == secondInk {
			t.Fatalf("the fill and its label share colour %v; the keyframe did not target the fill alone", secondFill)
		}
	})

	t.Run("path card recolours the path fill", func(t *testing.T) {
		if firstPathFillColour == basePathFillColour {
			t.Fatalf("the path fill stayed at %v", basePathFillColour)
		}
		if secondPathFillColour == basePathFillColour {
			t.Fatalf("the path fill stayed at %v after two ticks", basePathFillColour)
		}
		if firstPathStrokeColour != basePathStrokeColour || secondPathStrokeColour != basePathStrokeColour {
			t.Fatalf("the path border moved: base %v, first %v, second %v",
				basePathStrokeColour, firstPathStrokeColour, secondPathStrokeColour)
		}
		for _, op := range []*ownframe.DisplayOp{secondPathFill, secondPathStroke} {
			corner, ok := op.CornerPathValue()
			if !ok || corner == nil || len(corner.Cmds) == 0 {
				t.Fatal("a colour sample dropped the path command list")
			}
		}
	})

	t.Run("paused card does not advance", func(t *testing.T) {
		box := boxByID(t, p, "play-pause-paused")
		if len(box.Animations) == 0 || box.Animations[0].PlayState != "paused" {
			t.Fatalf("paused card animations = %+v, want a paused sample", box.Animations)
		}
		if secondPaused != firstPaused {
			t.Fatalf("paused card moved from %v to %v", firstPaused, secondPaused)
		}
		if secondPlaying <= firstPlaying {
			t.Fatalf("playing card stayed at %v across the second tick (first %v)", secondPlaying, firstPlaying)
		}
	})

	if stats.LastError != "" {
		t.Fatalf("motion bind failed after two ticks: %s", stats.LastError)
	}
}

// TestTransitionCardsMoveBetweenEndpoints toggles each shipped transition card
// through its button, then samples the paint mid-flight. Colour, opacity, and
// the matrix translation must sit strictly between the committed endpoints; a
// snap would land on one of them.
func TestTransitionCardsMoveBetweenEndpoints(t *testing.T) {
	p := newTestPage(t)
	ctx := context.Background()

	fromColour := colourOf(requireFill(t, p, "transition-colour-tile"))
	fromOpacity := requireFill(t, p, "transition-opacity-tile").Opacity()
	fromTranslate := translationX(t, p, "transition-transform-tile")

	clickID(t, p, "toggle-transition-colour")
	clickID(t, p, "toggle-transition-opacity")
	clickID(t, p, "toggle-transition-transform")

	toColour := colourOf(requireFill(t, p, "transition-colour-tile"))
	toOpacity := requireFill(t, p, "transition-opacity-tile").Opacity()
	toTranslate := translationX(t, p, "transition-transform-tile")

	for channel := range fromColour {
		if fromColour[channel] == toColour[channel] {
			t.Fatalf("colour channel %d did not change: %v", channel, fromColour[channel])
		}
	}
	if fromOpacity == toOpacity {
		t.Fatalf("opacity did not change: %v", fromOpacity)
	}
	if fromTranslate == toTranslate {
		t.Fatalf("translation did not change: %v", fromTranslate)
	}

	stats := p.Probe().Motion()
	if stats.TransitionsBound != 3 {
		t.Fatalf("bound transitions = %d, want 3 (error %q, skip %q)",
			stats.TransitionsBound, stats.TransitionError, stats.TransitionSkipReason)
	}
	if stats.TransitionError != "" {
		t.Fatalf("transition bind failed: %s", stats.TransitionError)
	}

	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("first transition tick failed: %v", err)
	}

	time.Sleep(transitionProbeDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second transition tick failed: %v", err)
	}

	colour := colourOf(requireFill(t, p, "transition-colour-tile"))
	for channel := range colour {
		if !between(colour[channel], fromColour[channel], toColour[channel]) {
			t.Fatalf("colour channel %d = %v, want between %v and %v",
				channel, colour[channel], fromColour[channel], toColour[channel])
		}
	}

	if opacity := requireFill(t, p, "transition-opacity-tile").Opacity(); !between(opacity, fromOpacity, toOpacity) {
		t.Fatalf("opacity = %v, want between %v and %v", opacity, fromOpacity, toOpacity)
	}

	if translate := translationX(t, p, "transition-transform-tile"); !between(translate, fromTranslate, toTranslate) {
		t.Fatalf("translation = %v, want between %v and %v", translate, fromTranslate, toTranslate)
	}

	if stats := p.Probe().Motion(); stats.TransitionError != "" {
		t.Fatalf("transition bind failed after samples: %s", stats.TransitionError)
	}
}

// TestWindowTransitionSwitchFreezesThePresentedPaint covers the public switch:
// it is on by default, turning it off releases the renderers without writing,
// the presented paint stays, and turning it back on lets the next style change
// capture.
func TestWindowTransitionSwitchFreezesThePresentedPaint(t *testing.T) {
	p := newTestPage(t)
	ctx := context.Background()

	if !p.Window().Transitions() {
		t.Fatal("the transition driver is off by default")
	}

	baseColour := colourOf(requireFill(t, p, "transition-colour-tile"))
	clickID(t, p, "toggle-transition-colour")
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("first transition tick failed: %v", err)
	}
	time.Sleep(transitionProbeDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second transition tick failed: %v", err)
	}

	presented := colourOf(requireFill(t, p, "transition-colour-tile"))
	if presented == baseColour {
		t.Fatalf("the transition did not leave the base paint %v", baseColour)
	}

	p.Window().SetTransitions(false)
	if p.Window().Transitions() {
		t.Fatal("SetTransitions(false) left the driver on")
	}
	if stats := p.Probe().Motion(); stats.TransitionsBound != 0 {
		t.Fatalf("live transitions after the switch = %d, want 0", stats.TransitionsBound)
	}

	animationBefore := translationX(t, p, "translate-tile")
	time.Sleep(secondDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("tick with transitions off failed: %v", err)
	}
	if frozen := colourOf(requireFill(t, p, "transition-colour-tile")); frozen != presented {
		t.Fatalf("frozen paint = %v, want the presented %v", frozen, presented)
	}
	if animationAfter := translationX(t, p, "translate-tile"); animationAfter <= animationBefore {
		t.Fatalf("the animation driver stopped with the transitions off: %v then %v",
			animationBefore, animationAfter)
	}

	p.Window().SetTransitions(true)
	if !p.Window().Transitions() {
		t.Fatal("SetTransitions(true) did not enable the driver")
	}

	clickID(t, p, "toggle-transition-colour")
	if stats := p.Probe().Motion(); stats.TransitionsBound != 1 {
		t.Fatalf("transitions after enabling = %d, want 1", stats.TransitionsBound)
	}

	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("first reversal tick failed: %v", err)
	}
	time.Sleep(transitionProbeDelay)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatalf("second reversal tick failed: %v", err)
	}

	mid := colourOf(requireFill(t, p, "transition-colour-tile"))
	for channel := range mid {
		if !between(mid[channel], presented[channel], baseColour[channel]) {
			t.Fatalf("reversal channel %d = %v, want between the frozen %v and the base %v",
				channel, mid[channel], presented[channel], baseColour[channel])
		}
	}
}
