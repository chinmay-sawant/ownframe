package frame

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

// t0 is the fixed clock every animation test steps against.
var t0 = time.Unix(0, 0)

// cardDisplay builds a 400x200 canvas with one 20x10-point card at (10,10).
func cardDisplay() (*ownframe.Display, *layout.DisplayOp) {
	d := &ownframe.Display{
		Width: 400, Height: 200,
		PointsPerPixel: 0.75, PixelPerPoint: 4.0 / 3,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 10, R: 1, G: 1, B: 1, Alpha: 1},
		},
		Order: []int{0},
	}

	return d, &d.Ops[0]
}

func TestTranslatePatchesXY(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := TranslateTo(d, []*ownframe.DisplayOp{op}, 30, 20, 100*time.Millisecond, Linear)

	if _, live := a.Step(t0); !live {
		t.Fatal("finished at the start")
	}

	if op.X != 10 || op.Y != 10 {
		t.Fatalf("first step moved to %v,%v", op.X, op.Y)
	}

	a.Step(t0.Add(50 * time.Millisecond))

	if op.X != 25 || op.Y != 20 {
		t.Fatalf("halfway to %v,%v", op.X, op.Y)
	}

	dmg, live := a.Step(t0.Add(100 * time.Millisecond))

	if live {
		t.Fatal("still live at the end")
	}

	if op.X != 40 || op.Y != 30 {
		t.Fatalf("ended at %v,%v", op.X, op.Y)
	}

	if dmg.Empty() {
		t.Fatal("no damage on the last step")
	}

	if a.Active() {
		t.Fatal("Active after the end")
	}

	if _, live := a.Step(t0.Add(200 * time.Millisecond)); live {
		t.Fatal("live after the end")
	}
}

func TestFadePatchesAlpha(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := FadeTo(d, []*ownframe.DisplayOp{op}, 0.25, 100*time.Millisecond, Linear)

	a.Step(t0)
	a.Step(t0.Add(50 * time.Millisecond))

	if math.Abs(op.Alpha-0.625) > 1e-9 {
		t.Fatalf("halfway alpha = %v", op.Alpha)
	}

	if _, live := a.Step(t0.Add(100 * time.Millisecond)); live {
		t.Fatal("still live at the end")
	}

	if math.Abs(op.Alpha-0.25) > 1e-9 {
		t.Fatalf("final alpha = %v", op.Alpha)
	}
}
