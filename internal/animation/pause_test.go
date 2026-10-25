package animation

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// slidePlayer builds a 0..100 x-translation player with the supplied timing.
func slidePlayer(t *testing.T, timing Timing) (*Player, *layout.DisplayOp, time.Time) {
	t.Helper()

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
		{Offset: 0, Move: true},
		{Offset: 1, X: 100, Move: true},
	}, timing)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(200, 0)
	p.Start(start)

	return p, op, start
}

// TestPauseHoldsSampledPoseAndResumeContinues is the pause acceptance case: a
// pause at a sampled time freezes the pose, and a resume after half a second
// continues from that phase instead of restarting at 0%.
func TestPauseHoldsSampledPoseAndResumeContinues(t *testing.T) {
	timing := Timing{Duration: time.Second, Iterations: 1, Fill: FillForwards}
	p, op, start := slidePlayer(t, timing)

	p.Step(start.Add(300 * time.Millisecond))
	if op.X != 30 {
		t.Fatalf("sample before pause x = %v, want 30", op.X)
	}

	p.Pause(start.Add(300 * time.Millisecond))
	if !p.Paused() {
		t.Fatal("Pause left the player running")
	}
	if p.Step(start.Add(800 * time.Millisecond)) {
		t.Fatal("a paused player reported active")
	}
	if op.X != 30 {
		t.Fatalf("paused sample x = %v, want the frozen 30", op.X)
	}

	// Half a second passed while paused. The resumed player must match an
	// uninterrupted run at the same active time, which is 400 ms.
	p.Resume(start.Add(800 * time.Millisecond))
	p.Step(start.Add(900 * time.Millisecond))
	if op.X != 40 {
		t.Fatalf("resumed sample x = %v, want 40 from 300 ms plus 100 ms", op.X)
	}

	uninterrupted, wantOp, wantStart := slidePlayer(t, timing)
	uninterrupted.Step(wantStart.Add(400 * time.Millisecond))
	if wantOp.X != op.X {
		t.Fatalf("resumed x = %v, uninterrupted x = %v at the same active time", op.X, wantOp.X)
	}
}

// TestPauseDuringDelay freezes a player that has not reached its active
// interval yet. The delay does not count while paused, so the resume continues
// the delay from where it stopped.
func TestPauseDuringDelay(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
		{Offset: 0, X: 100, Move: true},
		{Offset: 1, X: 200, Move: true},
	}, Timing{Duration: time.Second, Delay: 500 * time.Millisecond,
		Iterations: 1, Fill: FillBackwards})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(300, 0)
	p.Start(start)

	p.Step(start.Add(200 * time.Millisecond))
	if op.X != 100 {
		t.Fatalf("delay fill x = %v, want the backwards 100", op.X)
	}

	p.Pause(start.Add(200 * time.Millisecond))
	p.Step(start.Add(900 * time.Millisecond))
	if op.X != 100 {
		t.Fatalf("paused delay x = %v, want 100", op.X)
	}

	p.Resume(start.Add(900 * time.Millisecond))
	p.Step(start.Add(time.Second))
	if op.X != 100 {
		t.Fatalf("resumed delay x = %v, want 100 after 300 ms of delay", op.X)
	}
	p.Step(start.Add(1700 * time.Millisecond))
	if op.X != 150 {
		t.Fatalf("resumed midpoint x = %v, want 150 at 500 ms of active time", op.X)
	}
}

// TestClockCarriesAResumedPhase checks the second handoff: after a resume the
// clock is running again but keeps the active time from before the pause, so a
// later rebind must not drop that offset.
func TestClockCarriesAResumedPhase(t *testing.T) {
	timing := Timing{Duration: time.Second, Infinite: true}
	first, _, start := slidePlayer(t, timing)
	first.Step(start.Add(300 * time.Millisecond))
	first.Pause(start.Add(300 * time.Millisecond))
	first.Resume(start.Add(400 * time.Millisecond))
	first.Step(start.Add(500 * time.Millisecond))

	clock, ok := first.Clock()
	if !ok || clock.Paused {
		t.Fatalf("clock = %+v (ok %v), want a running clock", clock, ok)
	}

	second, secondOp, _ := slidePlayer(t, timing)
	second.SetClock(clock)
	second.Step(start.Add(600 * time.Millisecond))
	if secondOp.X != 50 {
		t.Fatalf("rebound running x = %v, want 50 at 500 ms of active time", secondOp.X)
	}
}

// TestClockCarriesAPausedPhaseAcrossPlayers mirrors the page handoff: a fresh
// player adopts the clock of the one it replaced and holds the same pose until
// it resumes.
func TestClockCarriesAPausedPhaseAcrossPlayers(t *testing.T) {
	timing := Timing{Duration: time.Second, Infinite: true}
	first, firstOp, start := slidePlayer(t, timing)
	first.Step(start.Add(250 * time.Millisecond))
	first.Pause(start.Add(250 * time.Millisecond))

	clock, ok := first.Clock()
	if !ok || !clock.Paused {
		t.Fatalf("clock = %+v (ok %v), want a paused clock", clock, ok)
	}

	second, secondOp, _ := slidePlayer(t, timing)
	second.SetClock(clock)
	if !second.Paused() {
		t.Fatal("SetClock did not pause the rebuilt player")
	}
	second.Step(start.Add(900 * time.Millisecond))
	if secondOp.X != firstOp.X {
		t.Fatalf("rebuilt paused x = %v, want %v", secondOp.X, firstOp.X)
	}

	second.Resume(start.Add(900 * time.Millisecond))
	second.Step(start.Add(950 * time.Millisecond))
	if secondOp.X != 30 {
		t.Fatalf("resumed rebuilt x = %v, want 30 at 300 ms of active time", secondOp.X)
	}
}

// TestPausedNegativeDelaySurvivesFiveRebinds is the drift case: a paused
// animation whose negative delay already advanced it a quarter of the way
// through its cycle is rebound above and over. The frozen sample must not
// move, and the clock must carry the negative-delay position itself rather
// than only the zero active time.
func TestPausedNegativeDelaySurvivesFiveRebinds(t *testing.T) {
	timing := Timing{Duration: 4 * time.Second, Delay: -time.Second,
		Infinite: true, Play: Paused}
	p, op, start := slidePlayer(t, timing)

	p.Step(start)
	if op.X != 25 {
		t.Fatalf("first paused sample x = %v, want 25 at a quarter of the cycle", op.X)
	}

	for i := 0; i < 5; i++ {
		clock, ok := p.Clock()
		if !ok || !clock.Paused {
			t.Fatalf("rebind %d clock = %+v (ok %v), want a paused clock", i, clock, ok)
		}
		if !clock.HasPosition || clock.Position != 0.25 {
			t.Fatalf("rebind %d position = %v (carried %v), want the sampled 0.25",
				i, clock.Position, clock.HasPosition)
		}

		next, nextOp, _ := slidePlayer(t, timing)
		next.SetClock(clock)
		next.Step(start.Add(time.Duration(i+1) * 50 * time.Millisecond))
		if nextOp.X != 25 {
			t.Fatalf("rebind %d x = %v, want the frozen 25", i, nextOp.X)
		}
		p, op = next, nextOp
	}
}

// TestPausedPositiveDelaySurvivesFiveRebinds covers the other sign: the
// animation is still inside its delay, so the paused sample is the backwards
// fill at the start keyframe and the carried position is negative.
func TestPausedPositiveDelaySurvivesFiveRebinds(t *testing.T) {
	timing := Timing{Duration: 4 * time.Second, Delay: time.Second,
		Infinite: true, Fill: FillBoth, Play: Paused}
	p, op, start := slidePlayer(t, timing)

	p.Step(start)
	if op.X != 0 {
		t.Fatalf("paused delay sample x = %v, want the backwards fill 0", op.X)
	}

	for i := 0; i < 5; i++ {
		clock, ok := p.Clock()
		if !ok || !clock.Paused || !clock.HasPosition {
			t.Fatalf("rebind %d clock = %+v (ok %v), want a paused position", i, clock, ok)
		}
		if clock.Position != -0.25 {
			t.Fatalf("rebind %d position = %v, want -0.25 inside the delay",
				i, clock.Position)
		}

		next, nextOp, _ := slidePlayer(t, timing)
		next.SetClock(clock)
		next.Step(start.Add(time.Duration(i+1) * 50 * time.Millisecond))
		if nextOp.X != 0 {
			t.Fatalf("rebind %d x = %v, want the delayed 0", i, nextOp.X)
		}
		p, op = next, nextOp
	}
}

// TestPausedMidCycleAfterRestartSurvivesRebinds restarts the animation from a
// trigger, pauses it inside the new cycle, and rebinds five times. The frozen
// phase must follow the restart, not the original run.
func TestPausedMidCycleAfterRestartSurvivesRebinds(t *testing.T) {
	timing := Timing{Duration: 4 * time.Second, Infinite: true}
	p, op, start := slidePlayer(t, timing)

	p.Step(start.Add(3 * time.Second))
	if op.X != 75 {
		t.Fatalf("sample before the restart x = %v, want 75", op.X)
	}

	restart := start.Add(3 * time.Second)
	if err := p.TriggerEffect(0, "replay", restart); err != nil {
		t.Fatalf("replay: %v", err)
	}
	p.Step(restart.Add(250 * time.Millisecond))
	if op.X != 6.25 {
		t.Fatalf("sample after the restart x = %v, want 6.25", op.X)
	}
	p.Pause(restart.Add(250 * time.Millisecond))

	for i := 0; i < 5; i++ {
		clock, ok := p.Clock()
		if !ok || !clock.Paused || !clock.HasPosition {
			t.Fatalf("rebind %d clock = %+v (ok %v), want a paused position", i, clock, ok)
		}
		if clock.Position != 0.0625 {
			t.Fatalf("rebind %d position = %v, want the restarted 0.0625", i, clock.Position)
		}

		next, nextOp, _ := slidePlayer(t, timing)
		next.SetClock(clock)
		next.Step(restart.Add(time.Duration(i+1) * 50 * time.Millisecond))
		if nextOp.X != 6.25 {
			t.Fatalf("rebind %d x = %v, want the frozen 6.25", i, nextOp.X)
		}
		p, op = next, nextOp
	}
}

// TestResumeAfterPausedWallTimeWithNegativeDelay pauses a negative-delay
// animation, waits half a second, and resumes. The paused wall time must not
// count, so the sample matches an uninterrupted run at the same active time.
func TestResumeAfterPausedWallTimeWithNegativeDelay(t *testing.T) {
	timing := Timing{Duration: 4 * time.Second, Delay: -time.Second, Infinite: true}
	p, op, start := slidePlayer(t, timing)

	p.Step(start.Add(500 * time.Millisecond))
	if op.X != 37.5 {
		t.Fatalf("sample before the pause x = %v, want 37.5", op.X)
	}
	p.Pause(start.Add(500 * time.Millisecond))
	p.Step(start.Add(time.Second))
	if op.X != 37.5 {
		t.Fatalf("paused sample x = %v, want the frozen 37.5", op.X)
	}

	p.Resume(start.Add(time.Second))
	p.Step(start.Add(1100 * time.Millisecond))
	if op.X != 40 {
		t.Fatalf("resumed sample x = %v, want 40 after 600 ms of active time", op.X)
	}

	uninterrupted, wantOp, wantStart := slidePlayer(t, timing)
	uninterrupted.Step(wantStart.Add(600 * time.Millisecond))
	if wantOp.X != op.X {
		t.Fatalf("resumed x = %v, uninterrupted x = %v at the same active time", op.X, wantOp.X)
	}
}

// TestSetClockPinsTheCarriedPausedSample hands a rebuilt player a clock whose
// position is the only exact record of the paused phase. The new player has no
// start delay, so recomputing the phase from the carried active time would
// move the pose to the start keyframe.
func TestSetClockPinsTheCarriedPausedSample(t *testing.T) {
	first, _, start := slidePlayer(t, Timing{Duration: 4 * time.Second, Delay: -time.Second,
		Infinite: true, Play: Paused})
	first.Step(start)

	clock, ok := first.Clock()
	if !ok || !clock.HasPosition {
		t.Fatalf("clock = %+v (ok %v), want the sampled paused position", clock, ok)
	}

	second, op, _ := slidePlayer(t, Timing{Duration: 4 * time.Second,
		Infinite: true, Play: Paused})
	second.SetClock(clock)
	second.Step(start.Add(time.Second))
	if op.X != 25 {
		t.Fatalf("rebuilt x = %v, want the carried 25, not a recomputed 0", op.X)
	}
}
