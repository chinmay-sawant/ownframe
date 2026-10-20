package animation

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestAccumulateComposesCurrentValueOnLaterIteration(t *testing.T) {
	op := &ownframe.DisplayOp{Kind: ownframe.DisplayOpFillRect, X: 12, Alpha: 1}
	base := cssmotion.Read(op)
	p, err := New([]Target{{Op: op, From: base}}, []Keyframe{
		{Offset: 0, X: 4, Move: true}, {Offset: 1, X: 10, Move: true},
	}, Timing{Duration: time.Second, Iterations: 2, Fill: FillBoth, Composition: Accumulate})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(60, 0)
	p.Start(start)
	p.Step(start.Add(1500 * time.Millisecond))
	if want := base.X + 7; math.Abs(op.X-want) > .001 {
		t.Fatalf("accumulated X = %v, want underlying plus current value %v", op.X, want)
	}
}
