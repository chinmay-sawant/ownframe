package alert

import "testing"

func TestObserveRequiresFreshCoverageAndTracking(t *testing.T) {
	var a Analyzer
	clear := Observation{NearestMM: 4200, Coverage: 60, Tracking: true}
	for i := 0; i < ClearFramesRequired-1; i++ {
		if got := a.Observe(clear); got != Unknown {
			t.Fatalf("clear frame %d: got %v, want Unknown", i+1, got)
		}
	}
	if got := a.Observe(clear); got != Clear {
		t.Fatalf("final clear frame: got %v, want Clear", got)
	}
	for _, o := range []Observation{
		{NearestMM: 4200, Coverage: 0, Tracking: true},
		{NearestMM: 4200, Coverage: 60, Tracking: false},
		{NearestMM: 0, Coverage: 60, Tracking: true},
	} {
		if got := a.Observe(o); got != Unknown {
			t.Fatalf("Observe(%+v) = %v, want Unknown", o, got)
		}
	}
}

func TestObserveStopThresholdsTakeEffectImmediately(t *testing.T) {
	var a Analyzer
	for _, tc := range []struct {
		distance int32
		want     State
	}{
		{3000, Stop},
		{1200, StopNow},
	} {
		got := a.Observe(Observation{NearestMM: tc.distance, Coverage: 60, Tracking: true})
		if got != tc.want {
			t.Fatalf("distance %d: got %v, want %v", tc.distance, got, tc.want)
		}
	}
}

func TestObserveKeepsStopUntilRepeatedClear(t *testing.T) {
	var a Analyzer
	near := Observation{NearestMM: 2000, Coverage: 60, Tracking: true}
	if got := a.Observe(near); got != Stop {
		t.Fatalf("near result = %v, want Stop", got)
	}
	clear := Observation{NearestMM: 4200, Coverage: 60, Tracking: true}
	for i := 0; i < ClearFramesRequired-1; i++ {
		if got := a.Observe(clear); got != Stop {
			t.Fatalf("clear frame %d = %v, want Stop", i+1, got)
		}
	}
	if got := a.Observe(clear); got != Clear {
		t.Fatalf("final clear result = %v, want Clear", got)
	}
}
