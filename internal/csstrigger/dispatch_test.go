package csstrigger

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestDispatchRunsEntryAndExitActions(t *testing.T) {
	start := time.Unix(1, 0)
	op := &ownframe.DisplayOp{Kind: ownframe.DisplayOpFillRect, X: 4, Alpha: 1}
	base := cssmotion.Paint{X: 4, Alpha: 1, Opacity: 1}
	player, err := animation.NewComposite([]animation.Target{{Op: op, From: base}}, []animation.Effect{{
		Frames:  []animation.Keyframe{{Offset: 0, Move: true}, {Offset: 1, X: 100, Move: true}},
		Timing:  animation.Timing{Duration: time.Second, Iterations: 1, Fill: animation.FillBoth, TimingFunction: "linear"},
		Trigger: "--ready play reset",
	}})
	if err != nil {
		t.Fatal(err)
	}
	boxes := []ownframe.Box{{ParentIndex: -1}, {ParentIndex: -1,
		Animations: []layout.Animation{{Trigger: "--ready play reset"}}}}
	players := map[int]*animation.Player{1: player}
	player.Step(start)
	if err := Dispatch(boxes, Event{Name: "--ready", Source: 0, Active: true}, players, start); err != nil {
		t.Fatal(err)
	}
	player.Step(start.Add(500 * time.Millisecond))
	if op.X != 54 {
		t.Fatalf("entry frame x = %v, want 54", op.X)
	}
	if err := Dispatch(boxes, Event{Name: "--ready", Source: 0}, players, start.Add(500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	player.Step(start.Add(500 * time.Millisecond))
	if op.X != base.X {
		t.Fatalf("exit frame x = %v, want reset to %v", op.X, base.X)
	}
}
