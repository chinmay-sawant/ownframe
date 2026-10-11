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

func firstTextOp(display *layout.Display) *layout.DisplayOp {
	for i := range display.Ops {
		if display.Ops[i].Kind == layout.DisplayOpText && display.Ops[i].Font != nil {
			return &display.Ops[i]
		}
	}

	return nil
}
