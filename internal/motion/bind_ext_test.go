package motion_test

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/motion"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"math"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestOwnFrameAnimationRangeCenter(t *testing.T) {
	d, _, indices, paints, p, _ := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-timeline:scroll();animation-range-center:25%;animation-fill-mode:both;animation-timing-function:linear;`)
	p.StepWithInput(motion.FrameInput{ScrollProgress: .25})
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
		t.Errorf("center point moved %v px, want 5", got)
	}
	d, _, indices, paints, p, _ = runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-timeline:scroll();animation-range-center:source 50%;animation-fill-mode:both;animation-timing-function:linear;`)
	p.StepWithInput(motion.FrameInput{ScrollProgress: .3, SourceProgress: .1, TargetProgress: .5})
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
		t.Errorf("source/target center moved %v px, want 5", got)
	}
}

func TestOwnFrameAnimationComposition(t *testing.T) {
	base := `animation-name:ten,twenty;animation-duration:1s;animation-timing-function:linear;animation-fill-mode:both;`
	for _, c := range []struct {
		mode string
		want float64
	}{
		// An additive transform list concatenates onto the earlier effect, so
		// ten (0 to 10px) plus twenty (0 to 20px) moves 30px. A replacing list
		// still writes only the later effect's 20px.
		{"add,add", 30},
		{"replace,replace", 20},
	} {
		d, _, indices, paints, p, start := runtimePlayer(t, base+`animation-composition:`+c.mode+`;`)
		if p.EffectCount() != 2 {
			t.Fatalf("effects = %d, want 2", p.EffectCount())
		}
		p.Step(start.Add(1_000_000_000))
		got := movedX(d, indices, paints) / d.PointsPerPixel
		if math.Abs(got-c.want) > .02 {
			t.Errorf("composition %s moved %v px, want %v", c.mode, got, c.want)
		}
	}
	d, _, indices, paints, p, start := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-iteration-count:2;animation-composition:accumulate;animation-fill-mode:both;animation-timing-function:linear;`)
	p.Step(start.Add(1500_000_000))
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
		t.Errorf("accumulated animation moved %v px, want current effect value 5", got)
	}
}

// transformMatrixTolerance compares two matrices that came from the same CSS
// transform through different composition paths.
const transformMatrixTolerance = 1e-9

// laidTarget is one laid-out page and the box with id "target".
type laidTarget struct {
	display *layout.Display
	box     layout.Box
}

// layoutTarget lays out one page and returns the target box with its display.
func layoutTarget(t *testing.T, html, theme string) laidTarget {
	t.Helper()

	page, err := ownframe.New(ownframe.Config{HTML: html, Width: 200, Height: 120, Theme: theme})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := page.Probe().Display()
	for _, box := range page.Boxes() {
		if box.ID == "target" {
			return laidTarget{display: d, box: box}
		}
	}
	t.Fatal("fixture has no target box")

	return laidTarget{}
}

// sampleTargetEnd binds the target's animation and samples it at the end of a
// one-second duration. The matrix the operation carries is the animation's
// contribution, composed with the engine's parent and world metadata.
func sampleTargetEnd(t *testing.T, target laidTarget) layout.Matrix2D {
	t.Helper()

	player, indices, _, start := bindMatrixTarget(t, target)
	player.Step(start.Add(time.Second))

	return motion.Read(&target.display.Ops[indices[0]]).Matrix
}

// bindMatrixTarget captures and binds the target's declared animation.
func bindMatrixTarget(t *testing.T, target laidTarget) (*motion.Player, []int, []motion.Paint, time.Time) {
	t.Helper()

	indices := motion.OperationIndices(target.display, target.box)
	base := motion.Capture(target.display, target.box)
	if len(indices) == 0 || len(indices) != len(base) {
		t.Fatalf("fixture captured %d paints for %d operations", len(base), len(indices))
	}
	start := time.Unix(300, 0)
	player, err := motion.BindAnimation(target.display, target.box, indices, base, start)
	if err != nil {
		t.Fatal(err)
	}

	return player, indices, base, start
}

// requireMatrixNear compares an animated matrix with the matrix a fresh layout
// produced for the same transform. The expected value comes from
// layout.Box.Transform, never from a hand-copied constant.
func requireMatrixNear(t *testing.T, got, want layout.Matrix2D) {
	t.Helper()

	if math.Abs(got.A-want.A) > transformMatrixTolerance ||
		math.Abs(got.B-want.B) > transformMatrixTolerance ||
		math.Abs(got.C-want.C) > transformMatrixTolerance ||
		math.Abs(got.D-want.D) > transformMatrixTolerance ||
		math.Abs(got.E-want.E) > transformMatrixTolerance ||
		math.Abs(got.F-want.F) > transformMatrixTolerance {
		t.Fatalf("animated matrix = %+v, want the layout matrix %+v", got, want)
	}
}

const targetRules = `#target{width:40px;height:20px;background-color:rgb(20,40,60);`

func TestMatrixPathRotatesAboutTheTopLeftOrigin(t *testing.T) {
	animated := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`transform-origin:0 0;animation-name:turn;animation-duration:1s;`+
			`animation-fill-mode:both;animation-timing-function:linear;}`+
			`@keyframes turn{from{transform:rotate(0deg)}to{transform:rotate(90deg)}}`)

	got := sampleTargetEnd(t, animated)

	// The engine lays out the same declaration statically; its stamped World
	// is the expected matrix. transform-origin is the border-box top left.
	expected := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`transform-origin:0 0;transform:rotate(90deg);}`)
	requireMatrixNear(t, got, expected.box.Transform.World)

	// A 90-degree CSS rotation is y-down: the top-right corner maps to the
	// position below the top-left corner, which stays fixed.
	scale := animated.display.PointsPerPixel
	originX, originY := animated.box.X*scale, animated.box.Y*scale
	if x, y := got.Apply(originX, originY); math.Abs(x-originX) > transformMatrixTolerance ||
		math.Abs(y-originY) > transformMatrixTolerance {
		t.Fatalf("origin moved to %v,%v, want %v,%v", x, y, originX, originY)
	}
	rightX, rightY := (animated.box.X+animated.box.W)*scale, animated.box.Y*scale
	if x, y := got.Apply(rightX, rightY); math.Abs(x-originX) > transformMatrixTolerance ||
		math.Abs(y-(originY+animated.box.W*scale)) > transformMatrixTolerance {
		t.Fatalf("top-right corner maps to %v,%v, want %v,%v",
			x, y, originX, originY+animated.box.W*scale)
	}
}

func TestMatrixPathScalesUnderARotatedParent(t *testing.T) {
	const parentRules = `#parent{width:100px;height:60px;transform:rotate(30deg);transform-origin:50% 50%;}`

	animated := layoutTarget(t, `<div id="parent"><div id="target"></div></div>`,
		parentRules+targetRules+`transform-origin:0 0;animation-name:grow;animation-duration:1s;`+
			`animation-fill-mode:both;animation-timing-function:linear;}`+
			`@keyframes grow{from{transform:scale(1)}to{transform:scale(2)}}`)

	got := sampleTargetEnd(t, animated)

	expected := layoutTarget(t, `<div id="parent"><div id="target"></div></div>`,
		parentRules+targetRules+`transform-origin:0 0;transform:scale(2);}`)
	requireMatrixNear(t, got, expected.box.Transform.World)

	// A nested parent rotation must survive the scale: the matrix linear part
	// is a rotation composed with the scale, not a bare scale.
	if math.Abs(math.Hypot(got.A, got.B)-2) > transformMatrixTolerance ||
		math.Abs(math.Hypot(got.C, got.D)-2) > transformMatrixTolerance {
		t.Fatalf("scale 2 matrix = %+v, want both column lengths 2", got)
	}
}

func TestMatrixPathReplacesAStaticTransform(t *testing.T) {
	// The static scale is what the operation already carries at bind. The
	// keyframe must replace it, not multiply onto it.
	animated := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`transform-origin:0 0;transform:scale(1.5);`+
			`animation-name:turn;animation-duration:1s;animation-fill-mode:both;`+
			`animation-timing-function:linear;}`+
			`@keyframes turn{from{transform:rotate(0deg)}to{transform:rotate(90deg)}}`)

	got := sampleTargetEnd(t, animated)

	expected := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`transform-origin:0 0;transform:rotate(90deg);}`)
	requireMatrixNear(t, got, expected.box.Transform.World)

	// Doubling would leave the static 1.5 scale in the linear part.
	if math.Abs(math.Hypot(got.A, got.B)-1) > transformMatrixTolerance ||
		math.Abs(math.Hypot(got.C, got.D)-1) > transformMatrixTolerance {
		t.Fatalf("matrix = %+v kept the static scale", got)
	}
}

func TestMatrixPathAnimatesAnUntransformedChildOfATransformedParent(t *testing.T) {
	const parentRules = `#parent{width:100px;height:60px;transform:rotate(30deg);transform-origin:50% 50%;}`

	animated := layoutTarget(t, `<div id="parent"><div id="target"></div></div>`,
		parentRules+targetRules+`animation-name:slide;animation-duration:1s;`+
			`animation-fill-mode:both;animation-timing-function:linear;}`+
			`@keyframes slide{from{transform:translate(0px,0px)}to{transform:translate(20px,0px)}}`)

	// The box has no static transform; the engine still exports its parent,
	// world, and resolved default origin (the border-box centre).
	if animated.box.Transform.HasMatrix {
		t.Fatal("fixture target unexpectedly declares a static transform")
	}
	if animated.box.Transform.World != animated.box.Transform.Parent {
		t.Fatalf("untransformed world = %+v, want the parent %+v",
			animated.box.Transform.World, animated.box.Transform.Parent)
	}
	scale := animated.display.PointsPerPixel
	wantOriginX := (animated.box.X + animated.box.W/2) * scale
	wantOriginY := (animated.box.Y + animated.box.H/2) * scale
	if math.Abs(animated.box.Transform.OriginX-wantOriginX) > transformMatrixTolerance ||
		math.Abs(animated.box.Transform.OriginY-wantOriginY) > transformMatrixTolerance {
		t.Fatalf("origin = %v,%v, want the border-box centre %v,%v",
			animated.box.Transform.OriginX, animated.box.Transform.OriginY, wantOriginX, wantOriginY)
	}

	got := sampleTargetEnd(t, animated)

	expected := layoutTarget(t, `<div id="parent"><div id="target"></div></div>`,
		parentRules+targetRules+`transform:translate(20px,0px);}`)
	requireMatrixNear(t, got, expected.box.Transform.World)

	// A 20px translation after a 30-degree parent rotation moves diagonally,
	// so a wrong composition order is visible in E and F.
	parent := animated.box.Transform.Parent
	dx, dy := parent.Apply(20*scale, 0)
	if math.Abs(got.E-dx) > transformMatrixTolerance || math.Abs(got.F-dy) > transformMatrixTolerance {
		t.Fatalf("translation = %v,%v, want %v,%v in parent space", got.E, got.F, dx, dy)
	}
}

func TestMatrixPathLeavesAThreeDKeyframeForRestamp(t *testing.T) {
	// A static transform gives the operation a base matrix that the 2D path
	// must not overwrite when the keyframe is 3D.
	animated := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`transform-origin:0 0;transform:scale(1.5);`+
			`animation-name:flip;animation-duration:1s;animation-fill-mode:both;`+
			`animation-timing-function:linear;}`+
			`@keyframes flip{from{transform:rotateY(0deg)}to{transform:rotateY(90deg)}}`)

	player, indices, base, start := bindMatrixTarget(t, animated)
	player.Step(start.Add(500 * time.Millisecond))

	if got := motion.Read(&animated.display.Ops[indices[0]]).Matrix; got != base[0].Matrix {
		t.Fatalf("3D keyframe wrote matrix %+v, want the base %+v", got, base[0].Matrix)
	}
	if !player.RestampPending() {
		t.Fatal("a 3D keyframe was not counted for the restamp pass")
	}
}

func TestMatrixPathLeavesAThreeDElementForRestamp(t *testing.T) {
	// The element itself carries a 3D transform, so even a 2D keyframe cannot
	// be written as a 2D matrix over its projected operations.
	animated := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`transform:rotateY(20deg);`+
			`animation-name:slide;animation-duration:1s;animation-fill-mode:both;`+
			`animation-timing-function:linear;}`+
			`@keyframes slide{from{transform:translate(0px,0px)}to{transform:translate(20px,0px)}}`)

	if !animated.box.Transform.HasTransform3D {
		t.Fatal("the engine did not flag the 3D element")
	}

	player, indices, base, start := bindMatrixTarget(t, animated)
	player.Step(start.Add(500 * time.Millisecond))

	if got := motion.Read(&animated.display.Ops[indices[0]]).Matrix; got != base[0].Matrix {
		t.Fatalf("3D element wrote matrix %+v, want the base %+v", got, base[0].Matrix)
	}
	if !player.RestampPending() {
		t.Fatal("a 3D element was not counted for the restamp pass")
	}
}

// TestMismatchedThreeDimensionalKeyframesBindAndRestamp is the regression for
// the former bind error: rotateX(0) to rotateY(90) has different function
// lists, and both are 3D. CSS interpolates that pair through the four-by-four
// decomposition, so the bind must succeed and the sample must reach the
// engine's restamp pass instead of failing the page.
func TestMismatchedThreeDimensionalKeyframesBindAndRestamp(t *testing.T) {
	animated := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`animation-name:turn;animation-duration:1s;animation-fill-mode:both;`+
			`animation-timing-function:linear;}`+
			`@keyframes turn{from{transform:rotateX(0deg)}to{transform:rotateY(90deg)}}`)

	player, indices, base, start := bindMatrixTarget(t, animated)
	player.Step(start.Add(500 * time.Millisecond))

	if got := motion.Read(&animated.display.Ops[indices[0]]).Matrix; got != base[0].Matrix {
		t.Fatalf("mismatched 3D keyframe wrote matrix %+v, want the base %+v", got, base[0].Matrix)
	}
	if !player.RestampPending() {
		t.Fatal("a mismatched 3D keyframe did not reach the restamp pass")
	}
	values := player.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one", len(values))
	}
	if !strings.HasPrefix(values[0].Transform, "matrix3d(") {
		t.Fatalf("restamp transform = %q, want matrix3d()", values[0].Transform)
	}
}

// TestMismatchedMatrix3DKeyframesBindAndRestamp covers the matrix3d pair, which
// CSS also converts to four-by-four matrices before interpolating.
func TestMismatchedMatrix3DKeyframesBindAndRestamp(t *testing.T) {
	const identity = `matrix3d(1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1)`
	const quarterTurn = `matrix3d(0,0,-1,0,0,1,0,0,1,0,0,0,0,0,0,1)` // rotateY(90deg)
	animated := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`animation-name:turn;animation-duration:1s;animation-fill-mode:both;`+
			`animation-timing-function:linear;}`+
			`@keyframes turn{from{transform:`+identity+`}to{transform:`+quarterTurn+`}}`)

	player, _, _, start := bindMatrixTarget(t, animated)
	player.Step(start.Add(500 * time.Millisecond))

	if !player.RestampPending() {
		t.Fatal("a mismatched matrix3d pair did not reach the restamp pass")
	}
}

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
func sampleSteps(t *testing.T, d *layout.Display, boxIndex int) []motion.Bound {
	t.Helper()
	bounds, _, err := motion.BindBox(d, d.Boxes[0], boxIndex, started())
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
	if bounds[0].Mask != motion.MaskRGB {
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
	if bounds[0].Mask != motion.MaskMatrix {
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

	bounds, gaps, err := motion.BindBox(d, d.Boxes[0], -1, started())
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

// TestAnonymousBoxWithOpOwnersUsesOwnerScopedColor verifies that Runtime.Bind
// passes the box's display index even when box.ID is empty, so owner-scoped
// colour animations only mutate the box's own operations instead of its whole
// subtree.
func TestAnonymousBoxWithOpOwnersUsesOwnerScopedColor(t *testing.T) {
	d := cardDisplay(running("tint"))
	d.Boxes[0].ID = ""
	d.Keyframes = colorRule("tint", "background-color", "#ff0000", "#00ff00")

	r := motion.New()
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if gaps := r.Stats().Gaps; len(gaps) != 0 {
		t.Fatalf("gaps = %v, want none when OpOwners is populated", gaps)
	}
	wantR, wantG, wantB := d.Ops[2].R, d.Ops[2].G, d.Ops[2].B
	r.Step(d, started().Add(500*time.Millisecond))
	// Box 0 owns op 0; child box 1 owns op 2 (unchanged).
	if d.Ops[2].R != wantR || d.Ops[2].G != wantG || d.Ops[2].B != wantB {
		t.Fatalf("child fill mutated = %v,%v,%v, want unchanged %v,%v,%v", d.Ops[2].R, d.Ops[2].G, d.Ops[2].B, wantR, wantG, wantB)
	}
}

// TestCurrentColorResolvesFromTheTextOp checks that currentColor reads the
// element's own text colour, and that a box with no text op fails the bind.
func TestCurrentColorResolvesFromTheTextOp(t *testing.T) {
	d := cardDisplay(running("ink"))
	d.Ops[1].R, d.Ops[1].G, d.Ops[1].B = 0, 0, 1
	d.Keyframes = colorRule("ink", "color", "#000000", "currentColor")

	bounds, _, err := motion.BindBox(d, d.Boxes[0], 0, started())
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
	_, _, err = motion.BindBox(child, child.Boxes[1], 1, started())
	if !errors.Is(err, motion.ErrCurrentColor) {
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

	bounds, gaps, err := motion.BindBox(d, d.Boxes[0], 0, started())
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

// pathAnimationPage is a bevel-cornered box with a uniform border, the page the
// engine paints as two corner paths.
const pathAnimationPage = `<body style="margin:0;background:#ffffff">` +
	`<div id="shaped" style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
	`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
	`corner-shape:bevel"></div></body>`

// TestBackgroundRecoloursACornerPath checks the shaped contract end to end: a
// background-color keyframe binds to the corner path's fill op, the sample
// rewrites its colour, and the border path and the command list stay put.
func TestBackgroundRecoloursACornerPath(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), pathAnimationPage, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	boxIndex := -1

	for i := range display.Boxes {
		if display.Boxes[i].ID == "shaped" {
			boxIndex = i
		}
	}

	if boxIndex < 0 {
		t.Fatal("shaped box is missing from the display")
	}

	fill, stroke := -1, -1

	for i := range display.Ops {
		corner, ok := display.Ops[i].CornerPathValue()
		if display.Ops[i].Kind != render.OpPath || !ok || corner == nil {
			continue
		}

		if corner.Stroke {
			stroke = i
		} else {
			fill = i
		}
	}

	if fill < 0 || stroke < 0 {
		t.Fatalf("want a fill and a stroke path, got fill %d and stroke %d", fill, stroke)
	}

	display.Boxes[boxIndex].Animations = []layout.Animation{running("tint")}
	display.Keyframes = colorRule("tint", "background-color", "#ff0000", "#00ff00")

	bounds, gaps, err := motion.BindBox(display, display.Boxes[boxIndex], boxIndex, started())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}

	if len(gaps) != 0 {
		t.Fatalf("gaps = %v, want none", gaps)
	}

	if len(bounds) != 1 {
		t.Fatalf("players = %d, want one fill player", len(bounds))
	}

	if !slices.Equal(bounds[0].Indices, []int{fill}) {
		t.Fatalf("background targets = %v, want the path fill %d", bounds[0].Indices, fill)
	}

	if bounds[0].Mask != motion.MaskRGB {
		t.Fatalf("background mask = %b, want RGB only", bounds[0].Mask)
	}

	strokeR, strokeG, strokeB := display.Ops[stroke].R, display.Ops[stroke].G, display.Ops[stroke].B

	bounds[0].Player.Step(halfWay())

	if got := display.Ops[fill]; got.R != .5 || got.G != .5 || got.B != 0 {
		t.Fatalf("path fill = %v,%v,%v, want the sRGB midpoint .5,.5,0", got.R, got.G, got.B)
	}

	if got := display.Ops[stroke]; got.R != strokeR || got.G != strokeG || got.B != strokeB {
		t.Fatalf("path border = %v,%v,%v, want its own %v,%v,%v",
			got.R, got.G, got.B, strokeR, strokeG, strokeB)
	}

	corner, ok := display.Ops[fill].CornerPathValue()
	if !ok || corner == nil || len(corner.Cmds) == 0 {
		t.Fatal("the colour sample dropped the path command list")
	}
}

// TestStaticPerspectiveFeedsAPartialKeyframe checks the whole propagation path:
// Blinkless resolves `perspective: 800px` on the element, BindBox hands that
// base to the keyframe builder, and a list with only a 50% stop starts at the
// static camera instead of none.
func TestStaticPerspectiveFeedsAPartialKeyframe(t *testing.T) {
	page, err := ownframe.New(ownframe.Config{
		HTML: `<div id="target"></div>`, Width: 200, Height: 100,
		Theme: `#target{width:40px;height:20px;background-color:rgb(20,40,60);
			perspective:800px;perspective-origin:10px 20px;
			animation:camera 1s linear both}
			@keyframes camera{50%{perspective:500px}}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	d := page.Probe().Display()
	boxes := page.Boxes()
	index := -1
	var box ownframe.Box
	for i, candidate := range boxes {
		if candidate.ID == "target" {
			index, box = i, candidate
		}
	}
	if index < 0 {
		t.Fatal("the fixture has no target box")
	}
	if !box.Transform.HasPerspective {
		t.Fatalf("the engine did not report the static perspective: %+v", box.Transform)
	}

	start := time.Unix(300, 0)
	bounds, _, err := motion.BindBox(d, box, index, start)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 1 {
		t.Fatalf("players = %d, want one", len(bounds))
	}

	bounds[0].Player.Step(start)
	values := bounds[0].Player.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one", len(values))
	}
	if !values[0].HasPerspective || !nearValue(values[0].Perspective, box.Transform.Perspective) {
		t.Fatalf("0%% perspective = %+v, want the static %v points",
			values[0], box.Transform.Perspective)
	}
	if !nearValue(values[0].PerspectiveOriginX, box.Transform.PerspectiveOriginX) ||
		!nearValue(values[0].PerspectiveOriginY, box.Transform.PerspectiveOriginY) {
		t.Fatalf("0%% origin = %v,%v, want the static %v,%v",
			values[0].PerspectiveOriginX, values[0].PerspectiveOriginY,
			box.Transform.PerspectiveOriginX, box.Transform.PerspectiveOriginY)
	}

	bounds[0].Player.Step(start.Add(500 * time.Millisecond))
	values = bounds[0].Player.RestampValues()
	if len(values) != 1 || !nearValue(values[0].Perspective, 375) {
		t.Fatalf("50%% perspective = %+v, want the declared 500px as 375 points", values)
	}
}

func nearValue(got, want float64) bool { return math.Abs(got-want) <= 1e-9 }

func TestOwnFrameAnimationRange(t *testing.T) {
	assertRange(t, `animation-range:25% 75%;`, .5, 5)
}

func TestOwnFrameAnimationRangeStart(t *testing.T) {
	assertRange(t, `animation-range-start:25%;`, .625, 5)
}

func TestOwnFrameAnimationRangeEnd(t *testing.T) {
	assertRange(t, `animation-range-end:75%;`, .375, 5)
}

func TestOwnFrameAnimationRangeUsesHostNamedRanges(t *testing.T) {
	d, _, indices, paints, p, _ := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-timeline:scroll();animation-range:entry 0% exit 100%;animation-fill-mode:both;animation-timing-function:linear;`)
	p.StepWithInput(motion.FrameInput{ScrollProgress: .5, Ranges: map[string]motion.Range{
		"entry": {Start: .2, End: .4}, "exit": {Start: .6, End: .8},
	}})
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
		t.Errorf("named range moved %v px, want 5", got)
	}
}

func assertRange(t *testing.T, property string, progress, want float64) {
	t.Helper()
	d, _, indices, paints, p, _ := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-timeline:scroll();animation-fill-mode:both;animation-timing-function:linear;`+property)
	p.StepWithInput(motion.FrameInput{ScrollProgress: progress})
	if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-want) > .02 {
		t.Errorf("%s at progress %.2f moved %v px, want %v", property, progress, got, want)
	}
}

const routeFrames = `@keyframes ten{from{transform:translateX(0px)}to{transform:translateX(10px)}}
@keyframes twenty{from{transform:translateX(0px)}to{transform:translateX(20px)}}`

func runtimePlayer(t *testing.T, declarations string) (*ownframe.Display, ownframe.Box, []int, []motion.Paint, *motion.Player, time.Time) {
	t.Helper()
	page, err := ownframe.New(ownframe.Config{HTML: `<div id="target"></div>`, Width: 200, Height: 100,
		Theme: `#target{width:40px;height:20px;background-color:rgb(20,40,60);` + declarations + `}` + routeFrames})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := page.Probe().Display()
	var box ownframe.Box
	for _, candidate := range page.Boxes() {
		if candidate.ID == "target" {
			box = candidate
		}
	}
	indices, base := motion.OperationIndices(d, box), motion.Capture(d, box)
	if len(indices) == 0 || len(indices) != len(base) {
		t.Fatal("fixture has no target paint")
	}
	start := time.Unix(100, 0)
	p, err := motion.BindAnimation(d, box, indices, base, start)
	if err != nil {
		t.Fatal(err)
	}
	return d, box, indices, base, p, start
}

// movedX reports how far a sampled transform moved the operation along x. A
// transform-list keyframe writes the operation matrix, not the paint X.
func movedX(d *ownframe.Display, indices []int, base []motion.Paint) float64 {
	return motion.Read(&d.Ops[indices[0]]).Matrix.E - base[0].Matrix.E
}

func TestOwnFrameAnimationTimeline(t *testing.T) {
	cases := []struct {
		timeline string
		input    motion.FrameInput
	}{
		{"scroll()", motion.FrameInput{ScrollProgress: .5}},
		{"view()", motion.FrameInput{ViewProgress: .5}},
		{"--custom", motion.FrameInput{Timelines: map[string]float64{"--custom": .5}}},
	}
	for _, c := range cases {
		d, _, indices, paints, p, _ := runtimePlayer(t, `animation-name:ten;animation-duration:1s;animation-iteration-count:1;animation-fill-mode:both;animation-timing-function:linear;animation-timeline:`+c.timeline+`;`)
		p.StepWithInput(c.input)
		if got := movedX(d, indices, paints) / d.PointsPerPixel; math.Abs(got-5) > .02 {
			t.Errorf("timeline %s moved %v px at halfway, want 5", c.timeline, got)
		}
	}
}

func TestAnimationResolvesPercentageTranslateAtBind(t *testing.T) {
	page, err := ownframe.New(ownframe.Config{HTML: `<div id="target"></div>`, Width: 200, Height: 100,
		Theme: `#target{width:40px;height:20px;background-color:rgb(20,40,60);` +
			`animation-name:pct;animation-duration:1s;animation-fill-mode:both;animation-timing-function:linear;}` +
			`@keyframes pct{from{transform:translate(0%,0%)}to{transform:translate(50%,25%)}}`})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := page.Probe().Display()
	var box layout.Box
	for _, candidate := range page.Boxes() {
		if candidate.ID == "target" {
			box = candidate
		}
	}
	indices, base := motion.OperationIndices(d, box), motion.Capture(d, box)
	if len(indices) == 0 || len(indices) != len(base) {
		t.Fatal("fixture has no target paint")
	}
	start := time.Unix(200, 0)
	p, err := motion.BindAnimation(d, box, indices, base, start)
	if err != nil {
		t.Fatal(err)
	}
	p.Step(start.Add(time.Second))
	got := motion.Read(&d.Ops[indices[0]]).Matrix
	if moved := (got.E - base[0].Matrix.E) / d.PointsPerPixel; math.Abs(moved-20) > .02 {
		t.Errorf("50%% translate moved %v px, want 20", moved)
	}
	if moved := (got.F - base[0].Matrix.F) / d.PointsPerPixel; math.Abs(moved-5) > .02 {
		t.Errorf("25%% translate moved %v px, want 5", moved)
	}
}

func TestOwnFrameTriggeredAnimationWaitsForAction(t *testing.T) {
	d, _, indices, paints, p, start := runtimePlayer(t, `animation-name:ten,twenty;animation-duration:1s;animation-timing-function:linear;animation-composition:add,add;animation-fill-mode:both;animation-trigger:none,--fire play;`)
	p.Step(start.Add(time.Second))
	if got := movedX(d, indices, paints) / d.PointsPerPixel; got < 9.98 || got > 10.02 {
		t.Fatalf("pending triggered effect changed paint: %v px", got)
	}
	if err := p.TriggerEffect(1, "play", start.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	p.Step(start.Add(1500 * time.Millisecond))
	// Both effects are additive, so the activated twenty effect (10px at its
	// midpoint) concatenates after the finished ten effect's 10px.
	if got := movedX(d, indices, paints) / d.PointsPerPixel; got < 19.98 || got > 20.02 {
		t.Fatalf("activated second effect moved %v px, want 20", got)
	}
}

// pathOwnerPage is a bevel-cornered box with a uniform border. The engine paints
// it as two DisplayOpPath ops, one background fill and one border stroke.
const pathOwnerPage = `<body style="margin:0;background:#ffffff">` +
	`<div id="shaped" style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
	`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
	`corner-shape:bevel"></div></body>`

// TestSelectSeparatesPathFillAndStroke checks that a colour property reaches a
// corner-shape path: a background recolours the fill half and a border or
// outline recolours the stroke half, even though both carry DisplayOpPath.
func TestSelectSeparatesPathFillAndStroke(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), pathOwnerPage, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	boxIndex := -1

	for i := range display.Boxes {
		if display.Boxes[i].ID == "shaped" {
			boxIndex = i
		}
	}

	if boxIndex < 0 {
		t.Fatal("shaped box is missing from the display")
	}

	fill, stroke := -1, -1

	for i := range display.Ops {
		corner, ok := display.Ops[i].CornerPathValue()
		if display.Ops[i].Kind != render.OpPath || !ok || corner == nil {
			continue
		}

		if corner.Stroke {
			stroke = i
		} else {
			fill = i
		}
	}

	if fill < 0 || stroke < 0 {
		t.Fatalf("want a fill and a stroke path, got fill %d and stroke %d", fill, stroke)
	}

	box := display.Boxes[boxIndex]

	if got := motion.Select(display, box, boxIndex, motion.KindFill); !slices.Equal(got, []int{fill}) {
		t.Fatalf("background targets = %v, want the path fill %d", got, fill)
	}

	if got := motion.Select(display, box, boxIndex, motion.KindStroke); !slices.Equal(got, []int{stroke}) {
		t.Fatalf("border targets = %v, want the path stroke %d", got, stroke)
	}

	if got := motion.Select(display, box, boxIndex, motion.KindInk); len(got) != 0 {
		t.Fatalf("a shaped box with no text targeted ink %v", got)
	}

	if got := motion.Select(display, box, boxIndex, motion.KindSubtree); !slices.Contains(got, fill) || !slices.Contains(got, stroke) {
		t.Fatalf("subtree targets = %v, want both path halves", got)
	}
}

func TestResolveUsesCommaAlignedAnimationRecords(t *testing.T) {
	page, err := ownframe.New(ownframe.Config{HTML: `<div id="target"></div>`, Width: 100,
		Height: 80, Theme: `#target{width:20px;height:20px;animation-name:first,second;` +
			`animation-trigger:--open play, --close play reset;}`})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	boxes := page.Boxes()
	for boxIndex, box := range boxes {
		if box.ID != "target" {
			continue
		}
		if len(box.Animations) != 2 || box.Animations[0].Trigger != "--open play" ||
			box.Animations[1].Trigger != "--close play reset" {
			t.Fatalf("animation records = %#v", box.Animations)
		}
		got := motion.ResolveTriggersForTest(boxes, motion.TriggerEvent{Name: "--close", Source: boxIndex, Active: false})
		if len(got) != 1 || got[0].AnimationIndex != 1 || got[0].Value != "reset" {
			t.Fatalf("close actions = %#v, want reset on animation 1", got)
		}
		return
	}
	t.Fatal("Blinkless display has no target box")
}
