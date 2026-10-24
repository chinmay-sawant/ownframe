package animation

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// linearPlayer builds a 0..100 x-translation player with a one-second cycle.
func linearPlayer(t *testing.T, timing Timing) (*Player, *layout.DisplayOp, time.Time) {
	t.Helper()

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
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
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
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
	want := 100 * cssmotion.Ease("ease", 0.5)
	if math.Abs(op.X-want) > 1e-9 {
		t.Fatalf("steps(1, jump-none) midpoint x = %v, want the ease fallback %v", op.X, want)
	}
}

// TestPerKeyframeEasingChangesSegmentMidpoint checks that a keyframe timing
// function eases the interval that starts at that keyframe, not the whole
// animation and not the interval before it.
func TestPerKeyframeEasingChangesSegmentMidpoint(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
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
	want := 100 + 100*cssmotion.Ease("ease-in", 0.5)
	if math.Abs(op.X-want) > 1e-9 {
		t.Fatalf("second segment x = %v, want %v from the keyframe ease-in", op.X, want)
	}
	if op.X >= 150 {
		t.Fatalf("second segment x = %v, want an eased midpoint below the linear 150", op.X)
	}
}
