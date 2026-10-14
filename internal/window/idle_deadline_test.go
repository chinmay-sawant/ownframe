package window

import (
	"testing"
	"time"
)

// TestPacePageDeadlineSchedulesOneWake pins that a page deadline posts one
// frame at its time and no keep-alive in between.
func TestPacePageDeadlineSchedulesOneWake(t *testing.T) {
	base := time.Unix(8000, 0)
	p := newPacePolicy(base, time.Millisecond, 125*time.Millisecond)
	at := idleAt(t, &p, base).Add(paceStep)

	if act := p.step(at, false, false, at.Add(time.Second)); act != paceHold {
		t.Fatalf("declared idle frame = %v, want hold", act)
	}

	if p.due(at.Add(125 * time.Millisecond)) {
		t.Fatal("the keep-alive fired while a page deadline was pending")
	}

	if p.due(at.Add(999 * time.Millisecond)) {
		t.Fatal("posted before the page deadline")
	}

	if !p.due(at.Add(time.Second)) {
		t.Fatal("missed the page deadline")
	}

	if p.due(at.Add(time.Second)) {
		t.Fatal("posted twice for one page deadline")
	}
}

// TestPaceZeroDeadlineFallsBackToKeepAlive pins that a page that never asks
// keeps the keep-alive fallback.
func TestPaceZeroDeadlineFallsBackToKeepAlive(t *testing.T) {
	base := time.Unix(8500, 0)
	p := newPacePolicy(base, time.Millisecond, 100*time.Millisecond)
	idle := idleAt(t, &p, base)

	if p.due(idle.Add(99 * time.Millisecond)) {
		t.Fatal("posted before the keep-alive")
	}

	if !p.due(idle.Add(100 * time.Millisecond)) {
		t.Fatal("missed the keep-alive")
	}
}
