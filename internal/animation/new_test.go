package animation

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestNewSortsKeyframesAndRejectsInvalidTargets(t *testing.T) {
	op := &layout.DisplayOp{}
	p, err := New([]Target{{Op: op}}, []Keyframe{{Offset: 1}, {Offset: 0}}, Timing{})
	if err != nil || p.effects[0].Frames[0].Offset != 0 {
		t.Fatalf("New() = %v, %v; want sorted frames", p, err)
	}
	if _, err := New([]Target{{}}, nil, Timing{}); err == nil {
		t.Fatal("New accepted a nil display operation")
	}
	if _, err := New([]Target{{Op: op}}, []Keyframe{{Offset: 1.1}}, Timing{}); err == nil {
		t.Fatal("New accepted an out-of-range keyframe")
	}
}

func TestPausedPlayerHoldsItsFrozenPose(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 8, Alpha: 1}
	base := cssmotion.Read(op)
	p, err := New([]Target{{Op: op, From: base}}, []Keyframe{
		{Offset: 0, Move: true}, {Offset: 1, X: 90, Move: true},
	}, Timing{Duration: time.Second, Iterations: 1, Play: Paused})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(50, 0)
	if p.Step(start) || op.X != 8 {
		t.Fatal("a paused player changed paint or reported active")
	}
	if p.Step(start.Add(900*time.Millisecond)) || op.X != 8 {
		t.Fatal("a paused player advanced its pose")
	}
	p.Resume(start.Add(900 * time.Millisecond))
	if !p.Step(start.Add(time.Second)) {
		t.Fatal("a resumed player stopped reporting active")
	}
	if op.X != 17 {
		t.Fatalf("resumed x = %v, want 17 from the frozen 0%% pose after 100 ms", op.X)
	}
}
