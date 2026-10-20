package animation

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func triggeredPlayer(t *testing.T, timing Timing) (*Player, *ownframe.DisplayOp, time.Time) {
	t.Helper()
	op := &ownframe.DisplayOp{Kind: ownframe.DisplayOpFillRect, Alpha: 1}
	p, err := NewComposite([]Target{{Op: op, From: cssmotion.Read(op)}}, []Effect{{
		Frames: []Keyframe{{Offset: 0, X: 10, Move: true}, {Offset: 1, X: 110, Move: true}},
		Timing: timing, Trigger: "--fire play",
	}})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(90, 0)
	p.Step(start)
	return p, op, start
}

func TestPlayWhileRunningKeepsProgress(t *testing.T) {
	p, op, start := triggeredPlayer(t, Timing{Duration: time.Second, Iterations: 1})
	if err := p.TriggerEffect(0, "play", start); err != nil {
		t.Fatal(err)
	}
	p.Step(start.Add(500 * time.Millisecond))
	if err := p.TriggerEffect(0, "play", start.Add(500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	p.Step(start.Add(750 * time.Millisecond))
	if op.X != 85 {
		t.Fatalf("play while running sampled x=%v, want 85", op.X)
	}
}

func TestPlayBackwardsChangesPlaybackRate(t *testing.T) {
	p, op, start := triggeredPlayer(t, Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth})
	if err := p.TriggerEffect(0, "play-backwards", start); err != nil {
		t.Fatal(err)
	}
	p.Step(start)
	if op.X != 110 {
		t.Fatalf("reverse start x=%v, want 110", op.X)
	}
	p.Step(start.Add(500 * time.Millisecond))
	if op.X != 60 {
		t.Fatalf("reverse midpoint x=%v, want 60", op.X)
	}
}
