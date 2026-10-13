package window

import (
	"testing"
	"time"
)

func TestPaceKeepAliveSchedule(t *testing.T) {
	base := time.Unix(5000, 0)
	p := newPacePolicy(base, time.Millisecond, 100*time.Millisecond)

	if p.due(base.Add(time.Second)) {
		t.Fatal("the schedule must not run on the fast path")
	}

	idle := idleAt(t, &p, base)

	if p.due(idle.Add(99 * time.Millisecond)) {
		t.Fatal("posted the keep-alive before the interval")
	}

	if !p.due(idle.Add(100 * time.Millisecond)) {
		t.Fatal("missed the keep-alive at the interval")
	}

	if p.due(idle.Add(101 * time.Millisecond)) {
		t.Fatal("posted two keep-alives for one interval")
	}

	if !p.due(idle.Add(200 * time.Millisecond)) {
		t.Fatal("missed the next keep-alive")
	}
}

func TestPaceInputWakesTheFastPath(t *testing.T) {
	base := time.Unix(6000, 0)
	p := newPacePolicy(base, time.Millisecond, 125*time.Millisecond)
	at := idleAt(t, &p, base).Add(paceStep)

	if act := p.step(at, false, true); act != paceToFast {
		t.Fatalf("input frame = %v, want fast", act)
	}

	if !p.fast {
		t.Fatal("input left the policy idle")
	}
}

func TestPaceDirtyBurstsThenWakes(t *testing.T) {
	base := time.Unix(7000, 0)
	p := newPacePolicy(base, time.Millisecond, 125*time.Millisecond)
	at := idleAt(t, &p, base).Add(paceStep)

	if act := p.step(at, true, false); act != paceBurst {
		t.Fatalf("first dirty idle frame = %v, want burst", act)
	}

	if p.fast {
		t.Fatal("one dirty frame left the window idle")
	}

	at = at.Add(paceStep)
	if act := p.step(at, true, false); act != paceToFast {
		t.Fatalf("second dirty idle frame = %v, want fast", act)
	}

	if !p.fast {
		t.Fatal("sustained damage left the policy idle")
	}
}
