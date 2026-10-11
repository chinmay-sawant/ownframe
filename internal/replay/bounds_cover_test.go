package replay

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
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
