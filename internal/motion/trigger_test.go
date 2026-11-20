package motion

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"testing"
	"time"
)

// splitTriggerDisplay is one element whose transform and background colour
// bind as two players. A trigger has to reach the colour player even though
// that player holds a single effect while the box lists two animations.
func splitTriggerDisplay() *layout.Display {
	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{
			{
				ID: testOpID, Kind: layout.DisplayOpFillRect,
				X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize, R: 0, Alpha: 1,
			},
			{
				ID: testOpID + 1, Kind: layout.DisplayOpText,
				X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize, R: 1, Alpha: 1,
			},
		},
		OpOwners: []int{0, 1},
		Boxes: []layout.Box{
			{
				ID: testChipID, Tag: "div", ParentIndex: -1,
				X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize,
				OperationIDs: []uint64{testOpID, testOpID + 1},
				Animations: []layout.Animation{
					{
						Name: testAnimation, DurationMS: testDurationMS,
						TimingFunction: "linear", IterationCount: 1, FillMode: "both",
						Trigger: "--move play",
					},
					{
						Name: testFade, DurationMS: testDurationMS,
						TimingFunction: "linear", IterationCount: 1, FillMode: "both",
						Trigger: "--tint play",
					},
				},
			},
			{
				ID: "label", Tag: "span", ParentIndex: 0,
				X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize,
				OperationIDs: []uint64{testOpID + 1},
			},
		},
		Keyframes: []layout.KeyframesRule{
			{Name: testAnimation, Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: testTransform, Value: testStartValue},
				}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: testTransform, Value: testShiftValue},
				}},
			}},
			{Name: testFade, Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "background-color", Value: "rgb(0,0,0)"},
				}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "background-color", Value: "rgb(255,255,255)"},
				}},
			}},
		},
	}
}

// TestTriggerReachesOnePlayerOfASplitElement proves the pass builder aligns a
// box animation index to the effect index of the player it actually built.
// The colour player holds one effect while the box lists two animations; a raw
// index would ask it for effect 1 and fail.
func TestTriggerReachesOnePlayerOfASplitElement(t *testing.T) {
	d := splitTriggerDisplay()
	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if got := r.Stats().Bound; got != 2 {
		t.Fatalf("bound players = %d, want two target sets", got)
	}

	passes := r.TriggerPasses()
	if len(passes) != 2 {
		t.Fatalf("trigger passes = %d, want 2", len(passes))
	}
	now := testClock()
	for i := range passes {
		event := TriggerEvent{Name: "--tint", Source: 0, Active: true}
		if err := DispatchTrigger(passes[i].Boxes, event, passes[i].Players, now); err != nil {
			t.Fatalf("pass %d: %v", i, err)
		}
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	if got := movedBy(&d.Ops[0]); got != 0 {
		t.Fatalf("a colour trigger moved the transform by %v", got)
	}
	if got := d.Ops[0].R; got < 0.49 || got > 0.51 {
		t.Fatalf("colour sample R = %v, want 0.5", got)
	}
}

func TestResolveUsesLastMatchingTriggerRecord(t *testing.T) {
	boxes := []layout.Box{{ParentIndex: -1, Animations: []layout.Animation{{
		Trigger: "--ready play --other pause --ready reset",
	}}}}
	got := resolveTriggers(boxes, TriggerEvent{Name: "--ready", Source: 0, Active: true})
	if len(got) != 1 || got[0].Value != "reset" {
		t.Fatalf("entry actions = %#v, want last matching reset", got)
	}
	got = resolveTriggers(boxes, TriggerEvent{Name: "--ready", Source: 0})
	if len(got) != 0 {
		t.Fatalf("exit actions = %#v, want no action", got)
	}
}

func TestResolveIgnoresInvalidEventActionCount(t *testing.T) {
	boxes := []layout.Box{{ParentIndex: -1, Animations: []layout.Animation{{
		Trigger: "--ready play pause reset",
	}}}}
	if got := resolveTriggers(boxes, TriggerEvent{Name: "--ready", Source: 0, Active: true}); len(got) != 0 {
		t.Fatalf("actions = %#v, want no action for three event actions", got)
	}
}

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

func TestAnimationTriggerActionsRespectTriggerScope(t *testing.T) {
	boxes := []layout.Box{
		{ID: "scope", ParentIndex: -1, TriggerScope: "--ready"},
		{ID: "source", ParentIndex: 0},
		{ID: "inside", ParentIndex: 0, Animations: []layout.Animation{{Trigger: "--ready play pause"}}},
		{ID: "outside", ParentIndex: -1, Animations: []layout.Animation{{Trigger: "--ready play"}}},
	}
	got := resolveTriggers(boxes, TriggerEvent{Name: "--ready", Source: 1, Active: true})
	if len(got) != 1 || got[0].BoxIndex != 2 || got[0].Value != "play" {
		t.Fatalf("enter actions = %#v, want scoped play on box 2", got)
	}
	got = resolveTriggers(boxes, TriggerEvent{Name: "--ready", Source: 1})
	if len(got) != 1 || got[0].BoxIndex != 2 || got[0].Value != "pause" {
		t.Fatalf("exit actions = %#v, want scoped pause on box 2", got)
	}
}

func triggeredPlayer(t *testing.T, timing Timing) (*Player, *layout.DisplayOp, time.Time) {
	t.Helper()
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := NewComposite([]Target{{Op: op, From: Read(op)}}, []Effect{{
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

func TestPlayOnceUsesOneIteration(t *testing.T) {
	p, op, start := triggeredPlayer(t, Timing{Duration: time.Second, Iterations: 3, Fill: FillBoth})
	if err := p.TriggerEffect(0, "play-once", start); err != nil {
		t.Fatal(err)
	}
	if p.Step(start.Add(1500*time.Millisecond)) || op.X != 110 {
		t.Fatalf("play-once x=%v, active=%v; want final x=110 and finished", op.X, p.Active())
	}
}

func TestTriggerResetPausesOnFirstDirectedKeyframe(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 8, Alpha: 1}
	p, err := NewComposite([]Target{{Op: op, From: Read(op)}}, []Effect{{
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
