package pacing_test

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
)

const frameStep = 16 * time.Millisecond
const cleanFrames = 5

func TestPaceCleanRunIdles(t *testing.T) {
	base := time.Unix(1000, 0)
	p := pacing.New(base)
	at := base

	for i := 1; i < cleanFrames; i++ {
		at = at.Add(frameStep)
		if act := p.Step(at, false, false, time.Time{}); act != pacing.Hold {
			t.Fatalf("clean frame %d = %v, want hold", i, act)
		}
	}

	at = at.Add(frameStep)
	if act := p.Step(at, false, false, time.Time{}); act != pacing.ToIdle {
		t.Fatalf("frame %d = %v, want idle", cleanFrames, act)
	}

	if !p.Snapshot().Idle {
		t.Fatal("clean streak did not publish idle mode")
	}
}

func TestPaceGraceHoldsAfterInput(t *testing.T) {
	base := time.Unix(2000, 0)
	p := pacing.New(base)
	last := base

	p.Step(base, false, true, time.Time{})

	for at := base.Add(frameStep); at.Sub(base) < 250*time.Millisecond; at = at.Add(frameStep) {
		if act := p.Step(at, false, false, time.Time{}); act != pacing.Hold {
			t.Fatalf("at %v = %v, want hold inside the grace", at.Sub(base), act)
		}

		last = at
	}

	at := last.Add(frameStep)
	if act := p.Step(at, false, false, time.Time{}); act != pacing.ToIdle {
		t.Fatalf("at %v = %v, want idle at the grace", at.Sub(base), act)
	}
}

func TestPaceAlwaysDirtyNeverIdles(t *testing.T) {
	base := time.Unix(3000, 0)
	p := pacing.New(base)
	at := base

	for i := 0; i < 600; i++ {
		at = at.Add(frameStep)
		if act := p.Step(at, true, false, time.Time{}); act == pacing.ToIdle {
			t.Fatalf("dirty frame %d idled the window", i)
		}

		if p.Snapshot().Idle {
			t.Fatalf("dirty frame %d left the fast path", i)
		}
	}
}

func TestPaceBarCadenceStillIdles(t *testing.T) {
	base := time.Unix(4000, 0)
	p := pacing.New(base)
	at := base

	// The board's refresh bar paints one rect every sixth frame.
	for i := 0; i < 12; i++ {
		at = at.Add(frameStep)

		dirty := i%6 == 0
		if act := p.Step(at, dirty, false, time.Time{}); act == pacing.ToIdle {
			if dirty || i < cleanFrames {
				t.Fatalf("idle on frame %d, dirty = %v", i, dirty)
			}

			return
		}
	}

	t.Fatal("a 10 Hz progress bar never idled the window")
}

func TestPaceKeepAliveSchedule(t *testing.T) {
	base := time.Unix(5000, 0)
	p := pacing.New(base)

	if p.Due(base.Add(time.Second)) {
		t.Fatal("the schedule must not run on the fast path")
	}

	idle := idleAt(t, &p, base)

	if p.Due(idle.Add(124 * time.Millisecond)) {
		t.Fatal("posted the keep-alive before the interval")
	}

	if !p.Due(idle.Add(125 * time.Millisecond)) {
		t.Fatal("missed the keep-alive at the interval")
	}

	if p.Due(idle.Add(126 * time.Millisecond)) {
		t.Fatal("posted two keep-alives for one interval")
	}

	if !p.Due(idle.Add(250 * time.Millisecond)) {
		t.Fatal("missed the next keep-alive")
	}
}

func TestPaceInputWakesTheFastPath(t *testing.T) {
	base := time.Unix(6000, 0)
	p := pacing.New(base)
	at := idleAt(t, &p, base).Add(frameStep)

	if act := p.Step(at, false, true, time.Time{}); act != pacing.ToFast {
		t.Fatalf("input frame = %v, want fast", act)
	}

	if p.Snapshot().Idle {
		t.Fatal("input left the policy idle")
	}
}

func TestPaceDirtyBurstsThenWakes(t *testing.T) {
	base := time.Unix(7000, 0)
	p := pacing.New(base)
	at := idleAt(t, &p, base).Add(frameStep)

	if act := p.Step(at, true, false, time.Time{}); act != pacing.Burst {
		t.Fatalf("first dirty idle frame = %v, want burst", act)
	}

	if !p.Snapshot().Idle {
		t.Fatal("one dirty frame left the window idle")
	}

	at = at.Add(frameStep)
	if act := p.Step(at, true, false, time.Time{}); act != pacing.ToFast {
		t.Fatalf("second dirty idle frame = %v, want fast", act)
	}

	if p.Snapshot().Idle {
		t.Fatal("sustained damage left the policy idle")
	}
}

// TestPacePageDeadlineSchedulesOneWake pins that a page deadline posts one
// frame at its time and no keep-alive in between.
func TestPacePageDeadlineSchedulesOneWake(t *testing.T) {
	base := time.Unix(8000, 0)
	p := pacing.New(base)
	at := idleAt(t, &p, base).Add(frameStep)

	if act := p.Step(at, false, false, at.Add(time.Second)); act != pacing.Hold {
		t.Fatalf("declared idle frame = %v, want hold", act)
	}

	if p.Due(at.Add(125 * time.Millisecond)) {
		t.Fatal("the keep-alive fired while a page deadline was pending")
	}

	if p.Due(at.Add(999 * time.Millisecond)) {
		t.Fatal("posted before the page deadline")
	}

	if !p.Due(at.Add(time.Second)) {
		t.Fatal("missed the page deadline")
	}

	if p.Due(at.Add(time.Second)) {
		t.Fatal("posted twice for one page deadline")
	}
}

// TestPaceZeroDeadlineFallsBackToKeepAlive pins that a page that never asks
// keeps the keep-alive fallback.
func TestPaceZeroDeadlineFallsBackToKeepAlive(t *testing.T) {
	base := time.Unix(8500, 0)
	p := pacing.New(base)
	idle := idleAt(t, &p, base)

	if p.Due(idle.Add(124 * time.Millisecond)) {
		t.Fatal("posted before the keep-alive")
	}

	if !p.Due(idle.Add(125 * time.Millisecond)) {
		t.Fatal("missed the keep-alive")
	}
}

// TestPaceEveryFrameRequestKeepsDrawing pins that a page that asks for the next
// frame now on every tick never leaves the fast loop.
func TestPaceEveryFrameRequestKeepsDrawing(t *testing.T) {
	base := time.Unix(9000, 0)
	p := pacing.New(base)
	at := base

	for i := 0; i < 600; i++ {
		at = at.Add(frameStep)
		if act := p.Step(at, false, false, at); act == pacing.ToIdle {
			t.Fatalf("an every-frame page idled at frame %d", i)
		}
	}

	if p.Snapshot().Idle {
		t.Fatal("an every-frame page left the fast loop")
	}
}

// TestPaceImmediateRequestWakesIdle pins that an immediate request from an
// idle window returns to uncapped drawing at once.
func TestPaceImmediateRequestWakesIdle(t *testing.T) {
	base := time.Unix(9500, 0)
	p := pacing.New(base)
	at := idleAt(t, &p, base).Add(frameStep)

	if act := p.Step(at, false, false, at); act != pacing.ToFast {
		t.Fatalf("immediate page request = %v, want fast", act)
	}

	if p.Snapshot().Idle {
		t.Fatal("an immediate page request left the window idle")
	}
}

// TestPaceDeclaredDirtyStaysIdle pins that a page that declared its next
// frame is trusted: one dirty frame does not burst or return to uncapped drawing.
func TestPaceDeclaredDirtyStaysIdle(t *testing.T) {
	base := time.Unix(9750, 0)
	p := pacing.New(base)
	at := idleAt(t, &p, base).Add(frameStep)
	want := at.Add(125 * time.Millisecond)

	if act := p.Step(at, true, false, want); act != pacing.Hold {
		t.Fatalf("declared dirty frame = %v, want hold", act)
	}

	if !p.Snapshot().Idle {
		t.Fatal("a declared dirty frame left the window idle")
	}

	if !p.Due(want) {
		t.Fatal("the declared deadline did not post")
	}
}

// TestPaceNextReportsTheSchedule pins the source of the devtools deadline.
func TestPaceNextReportsTheSchedule(t *testing.T) {
	base := time.Unix(9800, 0)
	p := pacing.New(base)

	if !p.Snapshot().Next.IsZero() {
		t.Fatal("the fast path with no request has no deadline")
	}

	first := base.Add(frameStep)
	p.Step(first, false, false, first.Add(time.Second))

	if got := p.Snapshot().Next; !got.Equal(first.Add(time.Second)) {
		t.Fatalf("fast next = %v, want the page deadline", got)
	}

	idle := idleAt(t, &p, base)
	if got := p.Snapshot().Next; !got.Equal(idle.Add(125 * time.Millisecond)) {
		t.Fatalf("next = %v, want the keep-alive", got)
	}

	at := idle.Add(frameStep)
	p.Step(at, false, false, at.Add(time.Second))

	if got := p.Snapshot().Next; !got.Equal(at.Add(time.Second)) {
		t.Fatalf("next = %v, want the page deadline", got)
	}
}

// TestPaceSkipCountsIdleFrames pins the skipped-frame counter.
func TestPaceSkipCountsIdleFrames(t *testing.T) {
	base := time.Unix(9900, 0)
	p := pacing.New(base)
	at := idleAt(t, &p, base)

	p.Step(at.Add(time.Second), false, false, at.Add(2*time.Second))

	if want := uint64(time.Second/frameStep) - 1; p.Snapshot().Skipped != want {
		t.Fatalf("skipped = %d, want %d", p.Snapshot().Skipped, want)
	}
}

// idleAt steps clean frames until the policy idles and returns that time.
func idleAt(t *testing.T, p *pacing.Policy, from time.Time) time.Time {
	t.Helper()

	at := from
	for i := 0; i < 100; i++ {
		at = at.Add(frameStep)
		if p.Step(at, false, false, time.Time{}) == pacing.ToIdle {
			return at
		}
	}

	t.Fatal("policy never idled")
	return at
}

func TestWaitTracksReplacedAndCancelledDeadlines(t *testing.T) {
	base := time.Unix(11000, 0)
	p := pacing.New(base)
	if got := p.Wait(base); got != 125*time.Millisecond/2 {
		t.Fatalf("fast wait = %v", got)
	}
	at := idleAt(t, &p, base)
	if got := p.Wait(at); got != 125*time.Millisecond {
		t.Fatalf("idle wait = %v", got)
	}
	p.Step(at, false, false, at.Add(time.Second))
	if got := p.Wait(at); got != time.Second {
		t.Fatalf("declared wait = %v", got)
	}
	earlier := at.Add(20 * time.Millisecond)
	p.Step(at, false, false, earlier)
	if got := p.Wait(at); got != 20*time.Millisecond {
		t.Fatalf("replacement wait = %v", got)
	}
	if got := p.Wait(earlier); got != time.Millisecond {
		t.Fatalf("due wait must be clamped, got %v", got)
	}
	p.Step(at, false, false, time.Time{})
	if p.Due(earlier) {
		t.Fatal("cancelled deadline posted a frame")
	}
	if got := p.Wait(at); got != 125*time.Millisecond {
		t.Fatalf("cancelled deadline wait = %v", got)
	}
	if !p.Due(at.Add(125 * time.Millisecond)) {
		t.Fatal("cancelling deadline lost keep-alive")
	}
}

func TestSnapshotDoesNotConsumeSchedule(t *testing.T) {
	base := time.Unix(12000, 0)
	p := pacing.New(base)
	at := idleAt(t, &p, base)
	deadline := at.Add(time.Second)
	p.Step(at, false, false, deadline)
	first := p.Snapshot()
	if !first.Idle || !first.Next.Equal(deadline) || first.Presents != cleanFrames+1 {
		t.Fatalf("snapshot = %+v", first)
	}
	if second := p.Snapshot(); second != first {
		t.Fatalf("reading snapshot mutated state: %+v != %+v", second, first)
	}
	if !p.Due(deadline) || p.Due(deadline) {
		t.Fatal("snapshot changed exactly-once deadline consumption")
	}
	if next := p.Snapshot().Next; !next.Equal(deadline.Add(125 * time.Millisecond)) {
		t.Fatalf("consumed deadline next = %v", next)
	}
}
