package animation

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// TestMissingKeyframeValuesUseTheUnderlyingPaint checks that a property
// omitted from a keyframe is not sampled as zero. A color omitted at 100%
// must hold the underlying color, and a transform omitted there must leave
// the operation where it started.
func TestMissingKeyframeValuesUseTheUnderlyingPaint(t *testing.T) {
	op := &ownframe.DisplayOp{Kind: ownframe.DisplayOpFillRect, X: 40, Y: 12,
		R: .42, G: .62, B: .21, Alpha: 1}
	base := cssmotion.Read(op)
	p, err := New([]Target{{Op: op, From: base}}, []Keyframe{
		{Offset: 0, Move: true},
		{Offset: .5, Move: true, X: 120, R: 1, ColorSet: true},
		{Offset: 1},
	}, Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth,
		TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(70, 0)
	p.Start(start)
	p.Step(start.Add(3 * time.Second))
	if math.Abs(op.R-base.R) > .001 || math.Abs(op.G-base.G) > .001 || math.Abs(op.B-base.B) > .001 {
		t.Fatalf("final color = %v,%v,%v, want underlying %v,%v,%v", op.R, op.G, op.B, base.R, base.G, base.B)
	}
	if op.X != base.X || op.Y != base.Y {
		t.Fatalf("final position = %v,%v, want underlying %v,%v", op.X, op.Y, base.X, base.Y)
	}
}
