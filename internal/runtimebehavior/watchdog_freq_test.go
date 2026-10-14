package runtimebehavior

import (
	"testing"
	"time"
)

func TestWatchdogFrequencyEdge(t *testing.T) {
	w := Watchdog{}
	if w.Frequency(0) != 0 || w.Frequency(time.Second) != 0 {
		t.Fatal("initial frequency wrong")
	}
	now := time.Now()
	if w.Sample(now, time.Second, 0) {
		t.Fatal("limit zero should never trip")
	}
	if w.IsOver() {
		t.Fatal("should not be over")
	}
	if w.Count() != 1 {
		t.Fatal("count wrong")
	}
	if w.Frequency(time.Second) <= 0 {
		t.Fatal("frequency should be positive")
	}
	if w.Sample(now.Add(10*time.Millisecond), time.Second, 10) {
		t.Fatal("should not trip under limit")
	}
	if w.Count() != 2 {
		t.Fatal("count should increment")
	}
	w.Reset()
	if w.Count() != 0 || w.IsOver() || w.LastAt != (time.Time{}) {
		t.Fatal("reset failed")
	}
	if w.Frequency(time.Second) != 0 {
		t.Fatal("reset frequency wrong")
	}
}
