package animation

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestStepDelayAlternateAndEndFill(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 5, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
		{Offset: 0, Move: true}, {Offset: 1, X: 10, Move: true},
	}, Timing{Duration: time.Second, Delay: 500 * time.Millisecond,
		EndDelay: 500 * time.Millisecond, Iterations: 2, Direction: Alternate, Fill: FillBoth})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(30, 0)
	p.Start(start)
	if !p.Step(start.Add(250*time.Millisecond)) || op.X != 5 {
		t.Fatal("backwards fill should hold the starting position during delay")
	}
	p.Step(start.Add(1500 * time.Millisecond))
	if op.X != 15 {
		t.Fatalf("alternate second iteration starts at x=%v, want 15", op.X)
	}
	if !p.Step(start.Add(2500*time.Millisecond)) || op.X != 5 {
		t.Fatal("end delay should keep the final backwards position active")
	}
	if p.Step(start.Add(3*time.Second)) || op.X != 5 {
		t.Fatal("finished animation should retain its forwards fill")
	}
}

func TestStepRepeatsInfiniteAlternate(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
		{Offset: 0, Move: true}, {Offset: 1, X: 10, Move: true},
	}, Timing{Duration: time.Second, Infinite: true, Direction: Alternate})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(40, 0)
	p.Start(start)
	if !p.Step(start.Add(1250*time.Millisecond)) || op.X != 7.5 {
		t.Fatalf("second iteration x=%v, want 7.5", op.X)
	}
}
