package cssanimation_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Fixture ids: the parent card paints a fill and a label, and the child card
// paints one fill inside it.
const (
	parentFillID = 11
	parentTextID = 12
	childFillID  = 21
)

// cardDisplay builds a parent card with a label and a nested child card. The
// parent's animation list is the caller's; the child has none.
func cardDisplay(animations ...layout.Animation) *layout.Display {
	return &layout.Display{
		Width: 200, Height: 120, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{
			{ID: parentFillID, Kind: layout.DisplayOpFillRect,
				X: 10, Y: 10, W: 100, H: 60, R: .1, G: .2, B: .3, Alpha: 1},
			{ID: parentTextID, Kind: layout.DisplayOpText,
				X: 20, Y: 30, W: 40, H: 12, R: 1, G: 1, B: 1, Alpha: 1},
			{ID: childFillID, Kind: layout.DisplayOpFillRect,
				X: 30, Y: 40, W: 40, H: 20, R: .4, G: .5, B: .6, Alpha: 1},
		},
		OpOwners: []int{0, 0, 1},
		Boxes: []layout.Box{
			{ID: "parent", X: 10, Y: 10, W: 100, H: 60,
				OperationIDs: []uint64{parentFillID, parentTextID, childFillID},
				Animations:   animations, ParentIndex: -1},
			{ID: "child", X: 30, Y: 40, W: 40, H: 20,
				OperationIDs: []uint64{childFillID}, ParentIndex: 0},
		},
	}
}

// running is one 1 s animation record with both fill edges held.
func running(name string) layout.Animation {
	return layout.Animation{Name: name, DurationMS: 1000, IterationCount: 1,
		FillMode: "both", PlayState: "running"}
}

// colorRule builds one linear two-stop colour animation.
func colorRule(name, property, from, to string) []layout.KeyframesRule {
	return []layout.KeyframesRule{{Name: name, Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: from}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: to}}},
	}}}
}

// started is the epoch every bind in this file uses.
func started() time.Time { return time.Unix(100, 0) }

// halfWay is the time one step reaches 50% of a 1 s animation.
func halfWay() time.Time { return started().Add(500 * time.Millisecond) }

// sampleSteps binds the parent card and samples it halfway once.
func sampleSteps(t *testing.T, d *layout.Display, boxIndex int) []cssanimation.Bound {
	t.Helper()
	bounds, _, err := cssanimation.BindBox(d, d.Boxes[0], boxIndex, started())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	for i := range bounds {
		bounds[i].Player.Step(halfWay())
	}
	return bounds
}

// TestBackgroundTargetsOnlyTheOwnerFill is the label regression: a
// background-color keyframe writes the animated colour onto the element's own
// fill, not onto its label text and not onto a child card's fill.
func TestBackgroundTargetsOnlyTheOwnerFill(t *testing.T) {
	d := cardDisplay(running("tint"))
	d.Keyframes = colorRule("tint", "background-color", "#ff0000", "#00ff00")

	bounds := sampleSteps(t, d, 0)
	if len(bounds) != 1 {
		t.Fatalf("players = %d, want one fill player", len(bounds))
	}
	if !slices.Equal(bounds[0].Indices, []int{0}) {
		t.Fatalf("background targets = %v, want the parent fill only", bounds[0].Indices)
	}
	if bounds[0].Mask != cssmotion.MaskRGB {
		t.Fatalf("background mask = %b, want RGB only", bounds[0].Mask)
	}
	if d.Ops[0].R != .5 || d.Ops[0].G != .5 || d.Ops[0].B != 0 {
		t.Fatalf("parent fill = %v,%v,%v, want the sRGB midpoint .5,.5,0",
			d.Ops[0].R, d.Ops[0].G, d.Ops[0].B)
	}
	if d.Ops[1].R != 1 || d.Ops[1].G != 1 || d.Ops[1].B != 1 {
		t.Fatalf("label text = %v,%v,%v, want its own 1,1,1", d.Ops[1].R, d.Ops[1].G, d.Ops[1].B)
	}
	if d.Ops[2].R != .4 || d.Ops[2].G != .5 || d.Ops[2].B != .6 {
		t.Fatalf("child fill = %v,%v,%v, want its own .4,.5,.6", d.Ops[2].R, d.Ops[2].G, d.Ops[2].B)
	}
}

// TestColourTargetsOnlyTheOwnerText is the other half of the kind rule: a text
// colour never writes a fill.
func TestColourTargetsOnlyTheOwnerText(t *testing.T) {
	d := cardDisplay(running("ink"))
	d.Keyframes = colorRule("ink", "color", "#000000", "#ff0000")

	bounds := sampleSteps(t, d, 0)
	if len(bounds) != 1 {
		t.Fatalf("players = %d, want one text player", len(bounds))
	}
	if !slices.Equal(bounds[0].Indices, []int{1}) {
		t.Fatalf("text colour targets = %v, want the label only", bounds[0].Indices)
	}
	if d.Ops[0].R != .1 || d.Ops[0].G != .2 || d.Ops[0].B != .3 {
		t.Fatalf("parent fill = %v,%v,%v, want it untouched", d.Ops[0].R, d.Ops[0].G, d.Ops[0].B)
	}
	if d.Ops[1].R != .5 || d.Ops[1].G != 0 || d.Ops[1].B != 0 {
		t.Fatalf("label text = %v,%v,%v, want the red midpoint .5,0,0", d.Ops[1].R, d.Ops[1].G, d.Ops[1].B)
	}
}

// TestTransformTargetsTheWholeSubtree pins that a moved parent moves its
// children and its own label, not only its own fill.
func TestTransformTargetsTheWholeSubtree(t *testing.T) {
	d := cardDisplay(running("slide"))
	d.Keyframes = []layout.KeyframesRule{{Name: "slide", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translateX(0px)"}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translateX(100px)"}}},
	}}}

	bounds := sampleSteps(t, d, 0)
	if len(bounds) != 1 {
		t.Fatalf("players = %d, want one subtree player", len(bounds))
	}
	if !slices.Equal(bounds[0].Indices, []int{0, 1, 2}) {
		t.Fatalf("transform targets = %v, want the whole subtree", bounds[0].Indices)
	}
	if bounds[0].Mask != cssmotion.MaskMatrix {
		t.Fatalf("transform mask = %b, want the matrix only", bounds[0].Mask)
	}
	for i := range d.Ops {
		if !d.Ops[i].XformSet {
			t.Fatalf("operation %d did not receive the subtree transform", i)
		}
		if got := d.Ops[i].Transform().E; got != 50 {
			t.Fatalf("operation %d translation = %v, want 50", i, got)
		}
	}
}

// TestDisjointAnimationsCompose checks two animations on one element with
// different target sets: the fill ends up both moved and recoloured.
func TestDisjointAnimationsCompose(t *testing.T) {
	d := cardDisplay(running("slide"), running("tint"))
	d.Keyframes = append(
		[]layout.KeyframesRule{{Name: "slide", Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translateX(0px)"}}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translateX(100px)"}}},
		}}},
		colorRule("tint", "background-color", "#ff0000", "#00ff00")...)

	bounds := sampleSteps(t, d, 0)
	if len(bounds) != 2 {
		t.Fatalf("players = %d, want a transform player and a fill player", len(bounds))
	}
	if got := d.Ops[0].Transform().E; got != 50 {
		t.Fatalf("fill translation = %v, want 50", got)
	}
	if d.Ops[0].R != .5 || d.Ops[0].G != .5 {
		t.Fatalf("fill colour = %v,%v, want the animated .5,.5", d.Ops[0].R, d.Ops[0].G)
	}
	if d.Ops[1].R != 1 || d.Ops[1].G != 1 || d.Ops[1].B != 1 {
		t.Fatalf("label colour = %v,%v,%v, want its own 1,1,1", d.Ops[1].R, d.Ops[1].G, d.Ops[1].B)
	}
}

// TestOwnerFallbackReportsAGap checks the documented fallback: a box with no
// trusted display index keeps the property on its kind-filtered subtree and the
// binder reports it.
func TestOwnerFallbackReportsAGap(t *testing.T) {
	d := cardDisplay(running("tint"))
	d.Keyframes = colorRule("tint", "background-color", "#ff0000", "#00ff00")

	bounds, gaps, err := cssanimation.BindBox(d, d.Boxes[0], -1, started())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 1 || !slices.Equal(bounds[0].Indices, []int{0, 2}) {
		t.Fatalf("fallback targets = %v, want both subtree fills", bounds[0].Indices)
	}
	if len(gaps) != 1 {
		t.Fatalf("gaps = %v, want one owner-fallback gap", gaps)
	}
}

// TestCurrentColorResolvesFromTheTextOp checks that currentColor reads the
// element's own text colour, and that a box with no text op fails the bind.
func TestCurrentColorResolvesFromTheTextOp(t *testing.T) {
	d := cardDisplay(running("ink"))
	d.Ops[1].R, d.Ops[1].G, d.Ops[1].B = 0, 0, 1
	d.Keyframes = colorRule("ink", "color", "#000000", "currentColor")

	bounds, _, err := cssanimation.BindBox(d, d.Boxes[0], 0, started())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	bounds[0].Player.Step(halfWay())
	if d.Ops[1].B != .5 || d.Ops[1].R != 0 {
		t.Fatalf("currentColor midpoint = %v,%v,%v, want 0,0,.5", d.Ops[1].R, d.Ops[1].G, d.Ops[1].B)
	}

	child := cardDisplay(running("ink"))
	child.Boxes[1].Animations = []layout.Animation{running("ink")}
	child.Keyframes = colorRule("ink", "color", "#000000", "currentColor")
	_, _, err = cssanimation.BindBox(child, child.Boxes[1], 1, started())
	if !errors.Is(err, cssanimation.ErrCurrentColor) {
		t.Fatalf("a box with no text op bound currentColor: %v", err)
	}
}

// TestDeferredPropertyIsCounted pins that a property a later phase owns is left
// at base paint and reported once, not dropped silently.
func TestDeferredPropertyIsCounted(t *testing.T) {
	d := cardDisplay(running("blur"))
	d.Keyframes = []layout.KeyframesRule{{Name: "blur", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "filter", Value: "blur(0px)"}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "filter", Value: "blur(4px)"}}},
	}}}

	bounds, gaps, err := cssanimation.BindBox(d, d.Boxes[0], 0, started())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 0 {
		t.Fatalf("a deferred property bound %d players, want none", len(bounds))
	}
	if len(gaps) != 1 {
		t.Fatalf("gaps = %v, want one deferred-property gap", gaps)
	}
}
