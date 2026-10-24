package cssanimation_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

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

	return cssmotion.Read(&target.display.Ops[indices[0]]).Matrix
}

// bindMatrixTarget captures and binds the target's declared animation.
func bindMatrixTarget(t *testing.T, target laidTarget) (*animation.Player, []int, []cssmotion.Paint, time.Time) {
	t.Helper()

	indices := cssmotion.OperationIndices(target.display, target.box)
	base := cssmotion.Capture(target.display, target.box)
	if len(indices) == 0 || len(indices) != len(base) {
		t.Fatalf("fixture captured %d paints for %d operations", len(base), len(indices))
	}
	start := time.Unix(300, 0)
	player, err := cssanimation.Bind(target.display, target.box, indices, base, start)
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

	if got := cssmotion.Read(&animated.display.Ops[indices[0]]).Matrix; got != base[0].Matrix {
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

	if got := cssmotion.Read(&animated.display.Ops[indices[0]]).Matrix; got != base[0].Matrix {
		t.Fatalf("3D element wrote matrix %+v, want the base %+v", got, base[0].Matrix)
	}
	if !player.RestampPending() {
		t.Fatal("a 3D element was not counted for the restamp pass")
	}
}
