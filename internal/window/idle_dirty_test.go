package window

import (
	"testing"
	"time"
)

func TestPaceAlwaysDirtyNeverIdles(t *testing.T) {
	base := time.Unix(3000, 0)
	p := newPacePolicy(base, 250*time.Millisecond, 125*time.Millisecond)
	at := base

	for i := 0; i < 600; i++ {
		at = at.Add(paceStep)
		if act := p.step(at, true, false, time.Time{}); act == paceToIdle {
			t.Fatalf("dirty frame %d idled the window", i)
		}

		if !p.fast {
			t.Fatalf("dirty frame %d left the fast path", i)
		}
	}
}

func TestPaceBarCadenceStillIdles(t *testing.T) {
	base := time.Unix(4000, 0)
	p := newPacePolicy(base, time.Millisecond, 125*time.Millisecond)
	at := base

	// The board's refresh bar paints one rect every sixth frame.
	for i := 0; i < 12; i++ {
		at = at.Add(paceStep)

		dirty := i%6 == 0
		if act := p.step(at, dirty, false, time.Time{}); act == paceToIdle {
			if dirty || i < paceCleanRun {
				t.Fatalf("idle on frame %d, dirty = %v", i, dirty)
			}

			return
		}
	}

	t.Fatal("a 10 Hz progress bar never idled the window")
}
