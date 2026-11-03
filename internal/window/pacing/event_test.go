package pacing_test

import (
	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
	"testing"
	"time"
)

func TestEventDrivenIdleHasNoKeepAlive(t *testing.T) {
	base := time.Unix(13000, 0)
	p := pacing.New(base)
	p.SetEventDriven(true)
	at := idleAt(t, &p, base)
	if !p.Snapshot().Next.IsZero() || p.Wait(at) != 0 || p.Due(at.Add(time.Hour)) {
		t.Fatal("static event-driven page scheduled an idle frame")
	}
	deadline := at.Add(time.Second)
	p.Step(at, false, false, deadline)
	if p.Wait(at) != time.Second || !p.Due(deadline) || p.Due(deadline) {
		t.Fatal("event-driven page lost its exactly-once deadline")
	}
	if p.Wait(deadline) != 0 {
		t.Fatal("consumed deadline restarted polling")
	}
	p.SetEventDriven(false)
	if p.Wait(deadline) == 0 {
		t.Fatal("legacy callback lost polling")
	}
}
