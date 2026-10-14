package window

import (
	"testing"
	"time"
)

// TestPaceEveryFrameRequestKeepsVsync pins that a page that asks for the next
// frame now on every tick never leaves the fast loop.
func TestPaceEveryFrameRequestKeepsVsync(t *testing.T) {
	base := time.Unix(9000, 0)
	p := newPacePolicy(base, time.Millisecond, 125*time.Millisecond)
	at := base

	for i := 0; i < 600; i++ {
		at = at.Add(paceStep)
		if act := p.step(at, false, false, at); act == paceToIdle {
			t.Fatalf("an every-frame page idled at frame %d", i)
		}
	}

	if !p.fast {
		t.Fatal("an every-frame page left the fast loop")
	}
}

// TestPaceImmediateRequestWakesIdle pins that an immediate request from an
// idle window returns to vsync at once.
func TestPaceImmediateRequestWakesIdle(t *testing.T) {
	base := time.Unix(9500, 0)
	p := newPacePolicy(base, time.Millisecond, 125*time.Millisecond)
	at := idleAt(t, &p, base).Add(paceStep)

	if act := p.step(at, false, false, at); act != paceToFast {
		t.Fatalf("immediate page request = %v, want fast", act)
	}

	if !p.fast {
		t.Fatal("an immediate page request left the window idle")
	}
}
