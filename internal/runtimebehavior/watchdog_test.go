package runtimebehavior

import (
	"testing"
	"time"
)

func TestWatchdogSampling(t *testing.T) {
	w := Watchdog{}
	now := time.Now()
	if w.Sample(now, time.Second, 2) {
		t.Fatal("first sample over")
	}
	if w.Sample(now.Add(10*time.Millisecond), time.Second, 2) {
		t.Fatal("second sample over")
	}
	if !w.Sample(now.Add(20*time.Millisecond), time.Second, 2) {
		t.Fatal("third sample should exceed limit")
	}
	if !w.Sample(now.Add(30*time.Millisecond), time.Second, 2) {
		t.Fatal("fourth sample should still exceed limit")
	}
	w.Reset()
	if w.Count() != 0 {
		t.Fatal("reset failed")
	}
}

func TestObserverCoverage(t *testing.T) {
	seen := ""
	o := NewResizeObserver(func(k, id string) { seen = k + "/" + id })
	o.Observe("box")
	o.Notify("resize", "box")
	if seen != "resize/box" {
		t.Fatal("observer callback broken")
	}
	o.Unobserve("box")
	if len(o.Observed) != 0 {
		t.Fatal("unobserve broken")
	}
	vo := NewVisibilityObserver(nil)
	vo.Notify("show", "node")
	if vo.LastKind != "show" {
		t.Fatal("visibility observer broken")
	}
}
