package alert

import "testing"

func TestNearClearBoundaryRevokesClear(t *testing.T) {
	var a Analyzer
	clear := Observation{NearestMM: 4200, Coverage: 60, Tracking: true}
	for i := 0; i < ClearFramesRequired; i++ {
		a.Observe(clear)
	}
	if got := a.State(); got != Clear {
		t.Fatalf("state after clear frames = %v, want Clear", got)
	}
	near := Observation{NearestMM: 3200, Coverage: 60, Tracking: true}
	if got := a.Observe(near); got != Unknown {
		t.Fatalf("near-boundary result = %v, want Unknown", got)
	}
	if got := a.Observe(clear); got != Unknown {
		t.Fatalf("first clear frame after boundary = %v, want Unknown", got)
	}
}

func TestNearClearBoundaryKeepsStop(t *testing.T) {
	var a Analyzer
	near := Observation{NearestMM: 2000, Coverage: 60, Tracking: true}
	a.Observe(near)
	boundary := Observation{NearestMM: 3200, Coverage: 60, Tracking: true}
	if got := a.Observe(boundary); got != Stop {
		t.Fatalf("stop at boundary = %v, want Stop", got)
	}
}
