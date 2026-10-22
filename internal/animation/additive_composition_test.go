package animation

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// TestAdditiveMissingKeyframeAddsNothing checks that a keyframe missing a
// property contributes nothing under an additive composition. The
// underlying paint must survive at 0% and 100% instead of being added to
// itself and saturating.
func TestAdditiveMissingKeyframeAddsNothing(t *testing.T) {
	op := &ownframe.DisplayOp{Kind: ownframe.DisplayOpFillRect, X: 8, Y: 6,
		R: .2, G: .4, B: .6, Alpha: 1}
	op.SetPaintOpacity(.5)
	base := cssmotion.Read(op)
	p, err := New([]Target{{Op: op, From: base}}, []Keyframe{
		{Offset: .5, Move: true, X: 40, Opacity: .25, OpacitySet: true,
			R: .3, G: .3, B: .3, ColorSet: true},
	}, Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth,
		TimingFunction: "linear", Composition: Add})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(80, 0)
	p.Start(start)
	p.Step(start)
	checkAdditivePaint(t, op, base, "start")
	p.Step(start.Add(500 * time.Millisecond))
	mid := base
	mid.R, mid.G, mid.B = base.R+.3, base.G+.3, base.B+.3
	mid.Opacity = base.Opacity + .25
	checkAdditivePaint(t, op, mid, "mid")
	if op.X != base.X+40 {
		t.Fatalf("mid X = %v, want %v", op.X, base.X+40)
	}
	p.Step(start.Add(3 * time.Second))
	checkAdditivePaint(t, op, base, "end")
	if op.X != base.X || op.Y != base.Y {
		t.Fatalf("end position = %v,%v, want %v,%v", op.X, op.Y, base.X, base.Y)
	}
}

// checkAdditivePaint reports whether the op holds the wanted paint.
func checkAdditivePaint(t *testing.T, op *ownframe.DisplayOp, want cssmotion.Paint, when string) {
	t.Helper()
	if math.Abs(op.R-want.R) > .001 || math.Abs(op.G-want.G) > .001 ||
		math.Abs(op.B-want.B) > .001 {
		t.Fatalf("%s color = %v,%v,%v, want %v,%v,%v",
			when, op.R, op.G, op.B, want.R, want.G, want.B)
	}
	if math.Abs(op.Opacity()-want.Opacity) > .001 {
		t.Fatalf("%s opacity = %v, want %v", when, op.Opacity(), want.Opacity)
	}
}
