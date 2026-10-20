package animation

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestTriggerResetPausesOnFirstDirectedKeyframe(t *testing.T) {
	op := &ownframe.DisplayOp{Kind: ownframe.DisplayOpFillRect, X: 8, Alpha: 1}
	p, err := NewComposite([]Target{{Op: op, From: cssmotion.Read(op)}}, []Effect{{
		Frames: []Keyframe{{Offset: 0, X: 24, Move: true}, {Offset: 1, X: 90, Move: true}},
		Timing: Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth}, Trigger: "--show play reset",
	}})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(80, 0)
	p.Step(start)
	if err := p.TriggerEffect(0, "play", start); err != nil {
		t.Fatal(err)
	}
	p.Step(start.Add(500 * time.Millisecond))
	if op.X != 65 {
		t.Fatalf("playing x = %v, want 65", op.X)
	}
	resetAt := start.Add(500 * time.Millisecond)
	if err := p.TriggerEffect(0, "reset", resetAt); err != nil {
		t.Fatal(err)
	}
	if p.Step(resetAt) || op.X != 32 {
		t.Fatalf("reset x = %v, active = %v; want first keyframe 32 and paused", op.X, p.Active())
	}
	p.Step(resetAt.Add(time.Second))
	if op.X != 32 {
		t.Fatalf("paused reset moved to x=%v, want 32", op.X)
	}
}
