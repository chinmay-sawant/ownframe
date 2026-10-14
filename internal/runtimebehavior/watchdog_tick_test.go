package runtimebehavior

import (
	"testing"
	"time"
)

func TestWatchdogTickAndWindow(t *testing.T) {
	w0 := Watchdog{}
	w0.Tick(0)
	if w0.Over || w0.IsOver() {
		t.Fatal("limit zero should clear over")
	}
	w := Watchdog{}
	for i := 0; i < 2; i++ {
		if w.Tick(2) {
			t.Fatal("should not trip yet")
		}
	}
	if !w.Tick(2) {
		t.Fatal("should trip on third tick")
	}
	if !w.IsOver() {
		t.Fatal("is over wrong")
	}
	w.Reset()
	// Backwards time clamps.
	now := time.Now()
	w.Sample(now, time.Second, 2)
	if w.Sample(now.Add(-time.Second), time.Second, 2) {
		t.Fatal("backwards sample should not trip")
	}
	if w.LastAt.Before(now) {
		t.Fatal("last at should clamp")
	}
	// Window rollover resets count baseline.
	w2 := Watchdog{}
	base := time.Now()
	w2.Sample(base, time.Second, 2)
	if w2.Sample(base.Add(1100*time.Millisecond), time.Second, 2) {
		t.Fatal("rollover should not trip")
	}
	if w2.Count() != 1 {
		t.Fatal("rollover count wrong")
	}
	_ = time.Millisecond
}
