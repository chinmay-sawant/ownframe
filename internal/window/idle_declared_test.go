package window

import (
	"testing"
	"time"
)

// TestPaceDeclaredDirtyStaysIdle pins that a page that declared its next
// frame is trusted: one dirty frame does not burst or return to vsync.
func TestPaceDeclaredDirtyStaysIdle(t *testing.T) {
	base := time.Unix(9750, 0)
	p := newPacePolicy(base, time.Millisecond, 125*time.Millisecond)
	at := idleAt(t, &p, base).Add(paceStep)
	want := at.Add(100 * time.Millisecond)

	if act := p.step(at, true, false, want); act != paceHold {
		t.Fatalf("declared dirty frame = %v, want hold", act)
	}

	if p.fast {
		t.Fatal("a declared dirty frame left the window idle")
	}

	if !p.due(want) {
		t.Fatal("the declared deadline did not post")
	}
}

// TestPaceNextReportsTheSchedule pins the source of the devtools deadline.
func TestPaceNextReportsTheSchedule(t *testing.T) {
	base := time.Unix(9800, 0)
	p := newPacePolicy(base, time.Millisecond, 100*time.Millisecond)

	if !p.next().IsZero() {
		t.Fatal("the fast path with no request has no deadline")
	}

	first := base.Add(paceStep)
	p.step(first, false, false, first.Add(time.Second))

	if got := p.next(); !got.Equal(first.Add(time.Second)) {
		t.Fatalf("fast next = %v, want the page deadline", got)
	}

	idle := idleAt(t, &p, base)
	if got := p.next(); !got.Equal(idle.Add(100 * time.Millisecond)) {
		t.Fatalf("next = %v, want the keep-alive", got)
	}

	at := idle.Add(paceStep)
	p.step(at, false, false, at.Add(time.Second))

	if got := p.next(); !got.Equal(at.Add(time.Second)) {
		t.Fatalf("next = %v, want the page deadline", got)
	}
}

// TestPaceSkipCountsIdleFrames pins the skipped-frame counter.
func TestPaceSkipCountsIdleFrames(t *testing.T) {
	base := time.Unix(9900, 0)
	p := newPacePolicy(base, time.Millisecond, 125*time.Millisecond)
	at := idleAt(t, &p, base)

	p.step(at.Add(time.Second), false, false, at.Add(2*time.Second))

	if want := uint64(time.Second/paceStep) - 1; p.skipped != want {
		t.Fatalf("skipped = %d, want %d", p.skipped, want)
	}
}
