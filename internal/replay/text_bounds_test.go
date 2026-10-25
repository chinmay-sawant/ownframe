package replay

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestOpBoundsTextUsesAscent(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayListState(context.Background(),
		`<html><body><p>Hello</p></body></html>`, 200, 100, render.State{})
	if err != nil {
		t.Fatal(err)
	}

	op := firstTextOp(display)
	if op == nil {
		t.Fatal("no text op")
	}

	ppt := display.PointsPerPixel

	box, ok := opBounds(op, ppt)
	if !ok {
		t.Fatal("text op unbounded")
	}

	if box.Min.Y >= int(op.Y/ppt) {
		t.Fatalf("box %v starts at or below the baseline", box)
	}

	below := image.Rect(box.Min.X, box.Max.Y+2, box.Max.X, box.Max.Y+9)
	if opTouches(op, ppt, below) {
		t.Fatal("a rect below the line box touched the run")
	}

	above := image.Rect(box.Min.X, box.Min.Y-5, box.Max.X, box.Min.Y+1)
	if !opTouches(op, ppt, above) {
		t.Fatal("a rect on the ascent did not touch the run")
	}
}

// TestOpBoundsTextWithoutPayload checks that a hand-built text op, which never
// attached a payload, still bounds as its line box. The scan reads the matrix
// field directly, so the payload must be bound before that read.
func TestOpBoundsTextWithoutPayload(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{
		Kind: layout.DisplayOpText, X: 20, Y: 40, W: 60, H: 16, InkDescent: 4, Size: 12,
	}

	got, ok := opBounds(op, 1)
	if !ok {
		t.Fatal("text op unbounded")
	}

	want := image.Rect(19, 27, 81, 45)
	if got != want {
		t.Fatalf("bounds = %v, want %v", got, want)
	}
}

// TestOpBoundsProjectedTextMoves checks that the projection branch still wins
// for a run whose baseline the engine already projected: the matrix box is far
// from the plain line box, and the plain box is culled.
func TestOpBoundsProjectedTextMoves(t *testing.T) {
	t.Parallel()

	plain := &layout.DisplayOp{
		Kind: layout.DisplayOpText, X: 20, Y: 40, W: 60, H: 16, InkDescent: 4, Size: 12,
	}
	matrix := layout.Matrix2D{A: 2, D: 2}

	projected := *plain
	projected.SetXform(matrix)
	projected.XformSet = false
	projected.X, projected.Y = matrix.Apply(plain.X, plain.Y)

	if box, ok := opBounds(&projected, 1); !ok || box != image.Rect(39, 55, 161, 89) {
		t.Fatalf("projected bounds = %v, %v", box, ok)
	}

	if opTouches(&projected, 1, image.Rect(21, 27, 79, 43)) {
		t.Fatal("the plain line box touched a projected run")
	}

	if !opTouches(&projected, 1, image.Rect(41, 56, 119, 88)) {
		t.Fatal("the matrix box did not touch the projected run")
	}

	if box, ok := opBounds(plain, 1); !ok || box != image.Rect(19, 27, 81, 45) {
		t.Fatalf("plain bounds = %v, %v", box, ok)
	}
}

func firstTextOp(display *layout.Display) *layout.DisplayOp {
	for i := range display.Ops {
		if display.Ops[i].Kind == layout.DisplayOpText && display.Ops[i].Font != nil {
			return &display.Ops[i]
		}
	}

	return nil
}
