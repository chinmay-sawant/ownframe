package motion

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"testing"
	"time"
)

func TestNewSortsKeyframesAndRejectsInvalidTargets(t *testing.T) {
	op := &layout.DisplayOp{}
	p, err := newAnimationPlayer([]Target{{Op: op}}, []Keyframe{{Offset: 1}, {Offset: 0}}, Timing{})
	if err != nil || p.effects[0].Frames[0].Offset != 0 {
		t.Fatalf("newAnimationPlayer() = %v, %v; want sorted frames", p, err)
	}
	if _, err := newAnimationPlayer([]Target{{}}, nil, Timing{}); err == nil {
		t.Fatal("New accepted a nil display operation")
	}
	if _, err := newAnimationPlayer([]Target{{Op: op}}, []Keyframe{{Offset: 1.1}}, Timing{}); err == nil {
		t.Fatal("New accepted an out-of-range keyframe")
	}
}

func TestPausedPlayerHoldsItsFrozenPose(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 8, Alpha: 1}
	base := Read(op)
	p, err := newAnimationPlayer([]Target{{Op: op, From: base}}, []Keyframe{
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

// slidePlayer builds a 0..100 x-translation player with the supplied timing.
func slidePlayer(t *testing.T, timing Timing) (*Player, *layout.DisplayOp, time.Time) {
	t.Helper()

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
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
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
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

func TestStepDelayAlternateAndEndFill(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 5, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
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
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
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

// linearPlayer builds a 0..100 x-translation player with a one-second cycle.
func linearPlayer(t *testing.T, timing Timing) (*Player, *layout.DisplayOp, time.Time) {
	t.Helper()

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, Move: true},
		{Offset: 1, X: 100, Move: true},
	}, timing)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(400, 0)
	p.Start(start)

	return p, op, start
}

// TestEndDelayHoldsThenReleases checks the holding state: the end delay keeps
// the final value live even without fill, and the release happens when the
// delay expires, not when the active interval ends.
func TestEndDelayHoldsThenReleases(t *testing.T) {
	cases := []struct {
		name         string
		fill         FillMode
		wantAfterEnd float64
	}{
		{"none releases", FillNone, 0},
		{"backwards releases", FillBackwards, 0},
		{"forwards keeps", FillForwards, 100},
		{"both keeps", FillBoth, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, op, start := linearPlayer(t, Timing{Duration: time.Second, Iterations: 1,
				EndDelay: 200 * time.Millisecond, Fill: tc.fill})

			if !p.Step(start.Add(1100 * time.Millisecond)) {
				t.Fatal("the end delay did not keep the player live")
			}
			if op.X != 100 {
				t.Fatalf("end delay x = %v, want the held final 100", op.X)
			}

			if p.Step(start.Add(1250 * time.Millisecond)) {
				t.Fatal("the player stayed live past the end delay")
			}
			if op.X != tc.wantAfterEnd {
				t.Fatalf("after end delay x = %v, want %v", op.X, tc.wantAfterEnd)
			}
		})
	}
}

// TestDurationAutoAppliesEndValue resolves auto to a zero-length animation:
// the end value applies at once, and the fill mode decides whether it stays.
func TestDurationAutoAppliesEndValue(t *testing.T) {
	forwards, op, start := linearPlayer(t, Timing{DurationAuto: true, Iterations: 1, Fill: FillForwards})
	if forwards.Step(start) {
		t.Fatal("a zero-length animation reported active")
	}
	if op.X != 100 {
		t.Fatalf("auto duration x = %v, want the end value 100", op.X)
	}

	none, noneOp, noneStart := linearPlayer(t, Timing{DurationAuto: true, Iterations: 1, Fill: FillNone})
	none.Step(noneStart)
	if noneOp.X != 0 {
		t.Fatalf("auto duration without fill x = %v, want the base 0", noneOp.X)
	}
}

// TestAutoDurationHonorsStartDelay checks that auto is a zero-length active
// interval after the declared delay, not an immediate end value regardless of
// the delay.
func TestAutoDurationHonorsStartDelay(t *testing.T) {
	p, op, start := linearPlayer(t, Timing{DurationAuto: true, Delay: 200 * time.Millisecond,
		Iterations: 1, Fill: FillBackwards})
	p.Step(start.Add(100 * time.Millisecond))
	if op.X != 0 {
		t.Fatalf("before the delay x = %v, want the start keyframe 0", op.X)
	}
	p.Step(start.Add(250 * time.Millisecond))
	if op.X != 0 {
		t.Fatalf("auto without forwards fill x = %v, want the base 0", op.X)
	}
}

// TestFractionalIterationSamplesTail pins that 1.25 iterations ends a quarter
// of the way through the second cycle rather than at the next whole keyframe.
func TestFractionalIterationSamplesTail(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, Move: true},
		{Offset: 0.5, X: 50, Move: true},
		{Offset: 1, X: 100, Move: true},
	}, Timing{Duration: time.Second, Iterations: 1.25, Fill: FillForwards})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(500, 0)
	p.Start(start)

	if !p.Step(start.Add(1200 * time.Millisecond)) {
		t.Fatal("the fractional tail ended early")
	}
	if math.Abs(op.X-20) > 1e-9 {
		t.Fatalf("tail x = %v, want 20 at a fifth of the second cycle", op.X)
	}

	if p.Step(start.Add(1250 * time.Millisecond)) {
		t.Fatal("the fractional animation stayed live past its count")
	}
	if op.X != 25 {
		t.Fatalf("finished fractional x = %v, want the held tail 25", op.X)
	}
}

// TestStepsJumpNoneHoldsEndValues checks the step positions that move the
// jumps inside the interval: jump-none starts at 0 and holds the final value
// early enough to end on it.
func TestStepsJumpNoneHoldsEndValues(t *testing.T) {
	p, op, start := linearPlayer(t, Timing{Duration: time.Second, Iterations: 1,
		Fill: FillForwards, TimingFunction: "steps(2, jump-none)"})

	for _, tc := range []struct {
		at   time.Duration
		want float64
	}{
		{250 * time.Millisecond, 0},
		{500 * time.Millisecond, 100},
		{750 * time.Millisecond, 100},
		{time.Second, 100},
	} {
		p.Step(start.Add(tc.at))
		if op.X != tc.want {
			t.Fatalf("steps(2, jump-none) at %v x = %v, want %v", tc.at, op.X, tc.want)
		}
	}
}

// TestStepsJumpBothHoldsEndValues checks jump-both: a jump at 0 and at 1, so
// the first plateau starts at 1/(n+1) and the last reaches the end value.
func TestStepsJumpBothHoldsEndValues(t *testing.T) {
	p, op, start := linearPlayer(t, Timing{Duration: time.Second, Iterations: 1,
		Fill: FillForwards, TimingFunction: "steps(2, jump-both)"})

	const part = 100.0 / 3.0
	for _, tc := range []struct {
		at   time.Duration
		want float64
	}{
		{100 * time.Millisecond, part},
		{500 * time.Millisecond, 2 * part},
		{900 * time.Millisecond, 2 * part},
		{time.Second, 100},
	} {
		p.Step(start.Add(tc.at))
		if math.Abs(op.X-tc.want) > 1e-9 {
			t.Fatalf("steps(2, jump-both) at %v x = %v, want %v", tc.at, op.X, tc.want)
		}
	}
}

// TestStepsOneJumpNoneIsInvalid checks the one step count jump-none rejects.
// The cascade drops that declaration, so a direct Timing falls back to the
// initial ease rather than pretending the function is linear.
func TestStepsOneJumpNoneIsInvalid(t *testing.T) {
	p, op, start := linearPlayer(t, Timing{Duration: time.Second, Iterations: 1,
		Fill: FillForwards, TimingFunction: "steps(1, jump-none)"})

	p.Step(start.Add(500 * time.Millisecond))
	want := 100 * Ease("ease", 0.5)
	if math.Abs(op.X-want) > 1e-9 {
		t.Fatalf("steps(1, jump-none) midpoint x = %v, want the ease fallback %v", op.X, want)
	}
}

// TestPerKeyframeEasingChangesSegmentMidpoint checks that a keyframe timing
// function eases the interval that starts at that keyframe, not the whole
// animation and not the interval before it.
func TestPerKeyframeEasingChangesSegmentMidpoint(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, X: 0, Move: true, TimingFunction: "linear"},
		{Offset: 0.5, X: 100, Move: true, TimingFunction: "ease-in"},
		{Offset: 1, X: 200, Move: true},
	}, Timing{Duration: time.Second, Iterations: 1, Fill: FillForwards})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(600, 0)
	p.Start(start)

	p.Step(start.Add(250 * time.Millisecond))
	if op.X != 50 {
		t.Fatalf("first segment x = %v, want the linear 50", op.X)
	}

	p.Step(start.Add(750 * time.Millisecond))
	want := 100 + 100*Ease("ease-in", 0.5)
	if math.Abs(op.X-want) > 1e-9 {
		t.Fatalf("second segment x = %v, want %v from the keyframe ease-in", op.X, want)
	}
	if op.X >= 150 {
		t.Fatalf("second segment x = %v, want an eased midpoint below the linear 150", op.X)
	}
}

func TestStepUsesSharedCSSTimingFunction(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
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
