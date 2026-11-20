package motion

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestDispatchRunsEntryAndExitActions(t *testing.T) {
	start := time.Unix(1, 0)
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Alpha: 1}
	base := Paint{X: 4, Alpha: 1, Opacity: 1}
	player, err := NewComposite([]Target{{Op: op, From: base}}, []Effect{{
		Frames:  []Keyframe{{Offset: 0, Move: true}, {Offset: 1, X: 100, Move: true}},
		Timing:  Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth, TimingFunction: "linear"},
		Trigger: "--ready play reset",
	}})
	if err != nil {
		t.Fatal(err)
	}
	boxes := []layout.Box{{ParentIndex: -1}, {ParentIndex: -1,
		Animations: []layout.Animation{{Trigger: "--ready play reset"}}}}
	players := map[int]*Player{1: player}
	player.Step(start)
	if err := DispatchTrigger(boxes, TriggerEvent{Name: "--ready", Source: 0, Active: true}, players, start); err != nil {
		t.Fatal(err)
	}
	player.Step(start.Add(500 * time.Millisecond))
	if op.X != 54 {
		t.Fatalf("entry frame x = %v, want 54", op.X)
	}
	if err := DispatchTrigger(boxes, TriggerEvent{Name: "--ready", Source: 0}, players, start.Add(500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	player.Step(start.Add(500 * time.Millisecond))
	if op.X != base.X {
		t.Fatalf("exit frame x = %v, want reset to %v", op.X, base.X)
	}
}
