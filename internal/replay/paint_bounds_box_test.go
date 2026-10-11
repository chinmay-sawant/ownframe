package replay

import (
	"context"
	"image"
	"math"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

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
