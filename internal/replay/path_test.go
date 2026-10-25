package replay

import (
	"context"
	"image"
	"math"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// cornerPathHTML is a bevel-cornered box with a uniform border. The engine
// emits the background fill and the border as two DisplayOpPath ops on the
// same border box.
func cornerPathHTML(radius int) string {
	return `<body style="margin:0;background:#ffffff">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:` +
		strconv.Itoa(radius) + `px;corner-shape:bevel"></div></body>`
}

// cornerPathRotatedHTML is a scoop-cornered box under a 15 degree rotation, so
// replay has to map every command point through the op matrix.
const cornerPathRotatedHTML = `<body style="margin:0;background:#ffffff">` +
	`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
	`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
	`corner-shape:scoop;transform:rotate(15deg)"></div></body>`

// pathFillOp returns the first fill path op of a display.
func pathFillOp(t *testing.T, display *layout.Display) *layout.DisplayOp {
	t.Helper()

	op := findPathOp(display, false)
	if op == nil {
		t.Fatal("no fill path op")
	}

	return op
}

// pathStrokeOp returns the first border path op of a display.
func pathStrokeOp(t *testing.T, display *layout.Display) *layout.DisplayOp {
	t.Helper()

	op := findPathOp(display, true)
	if op == nil {
		t.Fatal("no stroke path op")
	}

	return op
}

// TestPathBoundsCoverTheBorderBox pins the shaped box's painted bounds: a fill
// reaches the whole border box, and a border reaches half its width past it,
// each with the one-pixel replay slack.
func TestPathBoundsCoverTheBorderBox(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), cornerPathHTML(24), 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	ppt := display.PointsPerPixel

	fill := pathFillOp(t, display)

	box, ok := opBounds(fill, ppt)
	if !ok {
		t.Fatal("a path fill should be bounded")
	}

	if want := image.Rect(19, 19, 149, 109); box != want {
		t.Fatalf("fill bounds %v, want %v", box, want)
	}

	stroke := pathStrokeOp(t, display)

	box, ok = opBounds(stroke, ppt)
	if !ok {
		t.Fatal("a path border should be bounded")
	}

	if want := image.Rect(17, 17, 151, 111); box != want {
		t.Fatalf("border bounds %v, want %v", box, want)
	}
}

// TestPathBoundsFollowTheTransform checks that a transformed path bounds as
// its mapped command points, not as the unmapped box: the box grows, holds
// every mapped corner, and still rejects a rect away from the ink.
func TestPathBoundsFollowTheTransform(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), cornerPathRotatedHTML, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	ppt := display.PointsPerPixel
	op := pathFillOp(t, display)

	box, ok := opBounds(op, ppt)
	if !ok {
		t.Fatal("a transformed path should be bounded")
	}

	// A scoop is concave, so its rotated bounds are smaller than the rotated
	// box's, but a rotation must still push the paint past the upright box on
	// both axes.
	upright := boxBounds(op.X, op.Y, op.W, op.H, 0, ppt)
	if box.Min.X >= upright.Min.X || box.Min.Y >= upright.Min.Y ||
		box.Max.X <= upright.Max.X || box.Max.Y <= upright.Max.Y {
		t.Fatalf("rotated bounds %v did not grow past the upright box %v", box, upright)
	}

	m := op.Transform()
	centerX, centerY := m.Apply(op.X+op.W/2, op.Y+op.H/2)
	center := image.Pt(int(math.Round(centerX/ppt)), int(math.Round(centerY/ppt)))

	if !center.In(box) {
		t.Fatalf("mapped centre %v falls outside bounds %v", center, box)
	}

	if !opTouches(op, ppt, image.Rect(center.X-5, center.Y-5, center.X+5, center.Y+5)) {
		t.Fatal("a rect over the rotated box should touch")
	}

	if opTouches(op, ppt, image.Rect(180, 120, 190, 130)) {
		t.Fatal("a rect away from the rotated box should miss")
	}
}

// TestPathBoundsWithoutCommandsAreEmpty keeps a paint-free payload out of the
// dirty-region walk instead of drawing it everywhere.
func TestPathBoundsWithoutCommandsAreEmpty(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{Kind: layout.DisplayOpPath}

	box, ok := opBounds(op, 0.75)
	if !ok {
		t.Fatal("a path bounds as empty, not unbounded")
	}

	if !box.Empty() {
		t.Fatalf("bounds %v, want empty", box)
	}

	if _, _, _, _, _, present := pathInkBox(op); present {
		t.Fatal("a path without commands should report no ink box")
	}
}

// TestSigPathFoldsTheCommandList checks that two paths with the same box and
// colour but different corner geometry sign differently, while a redraw of the
// same page signs the same. Without this, a cached chunk would keep the old
// shape after a corner mutation.
func TestSigPathFoldsTheCommandList(t *testing.T) {
	t.Parallel()

	sharp, err := render.DisplayList(context.Background(), cornerPathHTML(24), 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	repeat, err := render.DisplayList(context.Background(), cornerPathHTML(24), 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	blunt, err := render.DisplayList(context.Background(), cornerPathHTML(16), 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	for _, stroke := range []bool{false, true} {
		pick := pathFillOp
		if stroke {
			pick = pathStrokeOp
		}

		a := sigOp(fnvOffsetBasis, pick(t, sharp))
		if again := sigOp(fnvOffsetBasis, pick(t, repeat)); a != again {
			t.Fatalf("stroke=%v: the same path signed differently", stroke)
		}

		if b := sigOp(fnvOffsetBasis, pick(t, blunt)); a == b {
			t.Fatalf("stroke=%v: different corner geometry signed the same", stroke)
		}
	}
}

// TestPathStrokeWidthScalesWithTheMatrix pins the stroke width rule: a
// rotation keeps the width and a uniform scale grows it, so a transformed
// border keeps its weight instead of thinning.
func TestPathStrokeWidthScalesWithTheMatrix(t *testing.T) {
	t.Parallel()

	rotated, err := render.DisplayList(context.Background(), cornerPathRotatedHTML, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	op := pathStrokeOp(t, rotated)
	if want := float32(op.Width * pxPerPt); pathStrokeWidth(op) != want {
		t.Fatalf("rotated width %v, want %v", pathStrokeWidth(op), want)
	}

	scaledPage := `<body style="margin:0;background:#ffffff">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
		`corner-shape:bevel;transform:scale(2)"></div></body>`

	scaled, err := render.DisplayList(context.Background(), scaledPage, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	op = pathStrokeOp(t, scaled)
	if want := float32(op.Width * pxPerPt * 2); math.Abs(float64(pathStrokeWidth(op)-want)) > 0.01 {
		t.Fatalf("scaled width %v, want %v", pathStrokeWidth(op), want)
	}
}
