package window

import (
	"testing"
	"time"
)

func TestPaceCleanRunIdles(t *testing.T) {
	base := time.Unix(1000, 0)
	p := newPacePolicy(base, time.Millisecond, 125*time.Millisecond)
	at := base

	for i := 1; i < paceCleanRun; i++ {
		at = at.Add(paceStep)
		if act := p.step(at, false, false); act != paceHold {
			t.Fatalf("clean frame %d = %v, want hold", i, act)
		}
	}

	at = at.Add(paceStep)
	if act := p.step(at, false, false); act != paceToIdle {
		t.Fatalf("frame %d = %v, want idle", paceCleanRun, act)
	}

	if p.fast || p.cleanRun < paceCleanRun {
		t.Fatalf("fast = %v, clean run = %d", p.fast, p.cleanRun)
	}
}

func TestPaceGraceHoldsAfterInput(t *testing.T) {
	base := time.Unix(2000, 0)
	p := newPacePolicy(base, 250*time.Millisecond, 125*time.Millisecond)
	last := base

	p.step(base, false, true)

	for at := base.Add(paceStep); at.Sub(base) < 250*time.Millisecond; at = at.Add(paceStep) {
		if act := p.step(at, false, false); act != paceHold {
			t.Fatalf("at %v = %v, want hold inside the grace", at.Sub(base), act)
		}

		last = at
	}

	at := last.Add(paceStep)
	if act := p.step(at, false, false); act != paceToIdle {
		t.Fatalf("at %v = %v, want idle at the grace", at.Sub(base), act)
	}
}
