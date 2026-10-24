package animation

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestStepUsesSharedCSSTimingFunction(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
		{Offset: 0, Move: true}, {Offset: 1, X: 10, Move: true},
	}, Timing{Duration: time.Second, Iterations: 1, TimingFunction: "steps(2, start)"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(25, 0)
	p.Start(start)
	p.Step(start.Add(250 * time.Millisecond))
	if op.X != 5 {
		t.Fatalf("x = %v, want 5 from CSS steps easing", op.X)
	}
}
