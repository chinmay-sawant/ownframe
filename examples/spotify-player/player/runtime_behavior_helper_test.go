package player

import (
	"testing"
	"time"
)

func TestRuntimeBehaviorHelper(t *testing.T) {
	h := NewRuntimeBehaviorHelper()
	h.StartLoading("spotify")
	if !h.Loading.Visible || h.Loading.Label != "spotify" {
		t.Fatal("loading start failed")
	}
	if !h.SlowLoading(time.Now().Add(time.Second), time.Millisecond) {
		t.Fatal("expected slow loading")
	}
	if !h.Marker.Slow {
		t.Fatal("marker alias out of sync")
	}
	h.ResolveLoading()
	if h.Loading.Visible {
		t.Fatal("loading not resolved")
	}

	h.Observe("track-row")
	h.Notify("select", "track-row")
	if h.Observer.LastKind != "select" || h.Observer.LastID != "track-row" {
		t.Fatal("observer notification failed")
	}
	if len(h.Visibility.Observed) != 1 {
		t.Fatal("visibility observe failed")
	}

	now := time.Now()
	if h.SampleWatchdog(now, time.Second, 1) {
		t.Fatal("unexpected first watchdog over")
	}
	if !h.SampleWatchdog(now.Add(10*time.Millisecond), time.Second, 1) {
		t.Fatal("expected watchdog over")
	}
	h.Reset()
	if h.Watchdog.Count() != 0 {
		t.Fatal("watchdog reset failed")
	}
}
