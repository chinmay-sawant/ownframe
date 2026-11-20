package replay

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"image"
	"math"
	"testing"
)

// TestBoundsCoverRectOps checks that the op filter never skips a rect-shaped
// op whose own box meets the rect. The nominal box is what a caller would
// hand the window as a dirty rect for that op.
func TestBoundsCoverRectOps(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayListState(context.Background(),
		`<html><body style="margin:0"><div style="width:100px;height:40px;`+
			`border:4px solid #123456"></div><p>text</p></body></html>`,
		320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}

	ppt := display.PointsPerPixel

	for i := range display.Ops {
		op := &display.Ops[i]
		switch op.Kind {
		case layout.DisplayOpFillRect, layout.DisplayOpStrokeRect, layout.DisplayOpImage:
		default:
			continue
		}

		nominal := image.Rect(
			int(op.X/ppt), int(op.Y/ppt),
			int((op.X+op.W)/ppt), int((op.Y+op.H)/ppt),
		)
		if nominal.Empty() {
			continue
		}

		if !opTouches(op, ppt, nominal) {
			t.Fatalf("op %d kind %d skipped its own box", i, op.Kind)
		}
	}
}

func TestClipRectClampsNotDrops(t *testing.T) {
	t.Parallel()

	display := &layout.Display{Width: 100, Height: 80, PointsPerPixel: 1}

	got := clipRect(display, image.Rect(-20, -20, 30, 30))
	if got != image.Rect(0, 0, 30, 30) {
		t.Fatalf("clamped = %v", got)
	}

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 10, H: 10}
	if !opTouches(op, 1, got) {
		t.Fatal("op at the canvas corner was dropped")
	}

	if !clipRect(display, image.Rect(-30, -30, -20, -20)).Empty() {
		t.Fatal("a rect fully outside should clamp away")
	}
}

func TestOpTouchesSkipsOutsideOps(t *testing.T) {
	t.Parallel()

	inside := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20}
	outside := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 80, Y: 80, W: 10, H: 10}
	rect := image.Rect(0, 0, 40, 40)

	if !opTouches(inside, 1, rect) {
		t.Fatal("inside op skipped")
	}

	if opTouches(outside, 1, rect) {
		t.Fatal("outside op drawn")
	}
}

func TestOpBoundsInertIsEmpty(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{Kind: layout.DisplayOpLinkURI, X: 1, Y: 1, W: 9, H: 9}

	box, ok := opBounds(op, 1)
	if !ok || !box.Empty() {
		t.Fatalf("link op bounds = %v, %v", box, ok)
	}
}

// overflowDisplay is a 100x100 canvas with one red fill below it, inside a
// 100x170 content. Points use 72/96 per pixel, so 75 points are 100 pixels.
func overflowDisplay() *layout.Display {
	const ppt = 72.0 / 96.0

	return &layout.Display{
		Width: 100, Height: 100, PointsPerPixel: ppt,
		Boxes: []layout.Box{{X: 0, Y: 150, W: 100, H: 20}},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 75, H: 75, R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 112.5, W: 75, H: 15, R: 1, Alpha: 1},
		},
		Order: []int{0, 1},
	}
}

// The buffer covers the content past the canvas, so a dirty rect below the
// canvas must survive clipping and still select its ops for replay.
func TestDrawRectKeepsOverflowRect(t *testing.T) {
	t.Parallel()

	display := overflowDisplay()
	rect := clipRect(display, image.Rect(0, 140, 100, 180))
	if rect != image.Rect(0, 140, 100, 170) {
		t.Fatalf("clipped = %v, want the content overlap", rect)
	}

	deep := &display.Ops[1]
	if !opTouches(deep, display.PointsPerPixel, rect) {
		t.Fatal("the fill below the canvas was skipped")
	}

	shallow := &display.Ops[0]
	if opTouches(shallow, display.PointsPerPixel, rect) {
		t.Fatal("the canvas fill leaks into the overflow rect")
	}
}

func TestClipRectDropsPastContent(t *testing.T) {
	t.Parallel()

	display := overflowDisplay()
	if !clipRect(display, image.Rect(0, 180, 100, 200)).Empty() {
		t.Fatal("a rect past the content should clamp away")
	}
}

// TestLineBoundsCoversDiagonal pins the stroked box of a diagonal segment:
// the union of both endpoints, each grown by half the stroke, so the far
// x+w and y+h extents stay inside. The old axis-aligned branch kept only
// the near endpoint's column.
func TestLineBoundsCoversDiagonal(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{
		Kind: layout.DisplayOpLine,
		X:    10, Y: 20, W: 60, H: 40, Width: 4,
	}

	padded, ok := opBounds(op, 1)
	if !ok || padded != image.Rect(7, 17, 73, 63) {
		t.Fatalf("padded = %v, %v", padded, ok)
	}

	got, ok := PaintBounds(op, 1)
	if !ok {
		t.Fatal("diagonal unbounded")
	}

	if got.Min.X > 8 || got.Min.Y > 18 || got.Max.X < 72 || got.Max.Y < 62 {
		t.Fatalf("diagonal bounds = %v, want far extents 72, 62", got)
	}
}

// TestPaintBoundsDropsTheReplaySlack pins the inspector's entry point: the
// same box the dirty-rect filter uses, without its one-pixel margin.
func TestPaintBoundsDropsTheReplaySlack(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 80, H: 40}

	padded, ok := opBounds(op, 1)
	if !ok || padded != image.Rect(9, 9, 91, 51) {
		t.Fatalf("padded = %v, %v", padded, ok)
	}

	paint, ok := PaintBounds(op, 1)
	if !ok || paint != image.Rect(10, 10, 90, 50) {
		t.Fatalf("paint = %v, %v", paint, ok)
	}
}

// TestPaintBoundsCoversStrokeAndLine checks that the painted box includes the
// ink outside the nominal geometry: half the stroke on each side of a border,
// and the stroke with its square cap around a rule.
func TestPaintBoundsCoversStrokeAndLine(t *testing.T) {
	t.Parallel()

	border := &layout.DisplayOp{Kind: layout.DisplayOpStrokeRect, X: 0, Y: 0, W: 80, H: 40, Width: 4}
	if got, ok := PaintBounds(border, 1); !ok || got != image.Rect(-2, -2, 82, 42) {
		t.Fatalf("stroke bounds = %v, %v", got, ok)
	}

	rule := &layout.DisplayOp{Kind: layout.DisplayOpLine, X: 0, Y: 10, W: 100, H: 0, Width: 4}
	if got, ok := PaintBounds(rule, 1); !ok || got != image.Rect(-2, 8, 102, 12) {
		t.Fatalf("line bounds = %v, %v", got, ok)
	}
}

// TestPaintBoundsMatchesElementBox places one background box and checks that
// the painted bounds agree with the element box the layout reports. The old
// inspector multiplied op points by PointsPerPixel, so the outline sat at 56%
// of the paint.
func TestPaintBoundsMatchesElementBox(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayListState(context.Background(),
		`<body style="margin:0"><div id="r" style="width:120px;height:40px;background:#1a56db"></div></body>`,
		320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}

	var box layout.Box

	for _, b := range display.Boxes {
		if b.ID == "r" {
			box = b
		}
	}

	if box.W <= 0 {
		t.Fatal("no #r box")
	}

	var fill *layout.DisplayOp

	for i := range display.Ops {
		op := &display.Ops[i]
		if op.Kind == layout.DisplayOpFillRect && colorNear(op, 0x1a, 0x56, 0xdb) {
			fill = op
		}
	}

	if fill == nil {
		t.Fatal("no blue fill")
	}

	got, ok := PaintBounds(fill, display.PointsPerPixel)
	if !ok {
		t.Fatal("fill unbounded")
	}

	want := image.Rect(int(box.X), int(box.Y), int(box.X+box.W), int(box.Y+box.H))
	if got != want {
		t.Fatalf("bounds = %v, want the element box %v", got, want)
	}
}

func colorNear(op *layout.DisplayOp, r, g, b float64) bool {
	const eps = 0.01

	return math.Abs(op.R-r/0xff) < eps &&
		math.Abs(op.G-g/0xff) < eps &&
		math.Abs(op.B-b/0xff) < eps
}
