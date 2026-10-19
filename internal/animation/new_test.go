package animation

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestNewSortsKeyframesAndRejectsInvalidTargets(t *testing.T) {
	op := &ownframe.DisplayOp{}
	p, err := New([]Target{{Op: op}}, []Keyframe{{Offset: 1}, {Offset: 0}}, Timing{})
	if err != nil || p.frames[0].Offset != 0 {
		t.Fatalf("New() = %v, %v; want sorted frames", p, err)
	}
	if _, err := New([]Target{{}}, nil, Timing{}); err == nil {
		t.Fatal("New accepted a nil display operation")
	}
	if _, err := New([]Target{{Op: op}}, []Keyframe{{Offset: 1.1}}, Timing{}); err == nil {
		t.Fatal("New accepted an out-of-range keyframe")
	}
}

func TestPausedPlayerDoesNotChangePaint(t *testing.T) {
	op := &ownframe.DisplayOp{Kind: ownframe.DisplayOpFillRect, X: 8, Alpha: 1}
	p, err := New([]Target{{Op: op}}, []Keyframe{{Offset: 1, X: 90, Move: true}},
		Timing{Duration: 1, Iterations: 1, Play: Paused})
	if err != nil {
		t.Fatal(err)
	}
	if p.Step(time.Unix(50, 0)) || op.X != 8 {
		t.Fatal("paused player changed paint or remained active")
	}
}
