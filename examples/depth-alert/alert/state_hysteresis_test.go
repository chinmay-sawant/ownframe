package alert

import "testing"

func TestUrgentStateSurvivesTransientDepthGap(t *testing.T) {
	for _, test := range []struct {
		name  string
		near  int32
		state State
	}{{"stop", 2000, Stop}, {"stop now", 800, StopNow}} {
		t.Run(test.name, func(t *testing.T) {
			var analyzer Analyzer
			got := analyzer.Observe(Observation{NearestMM: test.near, Coverage: 60, Tracking: true})
			if got != test.state {
				t.Fatalf("initial state = %v, want %v", got, test.state)
			}
			for i := 0; i < 3; i++ {
				got = analyzer.Observe(Observation{Tracking: true})
				if got != test.state {
					t.Fatalf("state after missing depth frame %d = %v, want %v", i+1, got, test.state)
				}
			}
		})
	}
}

func TestClearRequiresAllStableFrames(t *testing.T) {
	var analyzer Analyzer
	clear := Observation{NearestMM: 4200, Coverage: 60, Tracking: true}
	for i := 0; i < ClearFramesRequired-1; i++ {
		if got := analyzer.Observe(clear); got == Clear {
			t.Fatalf("state became Clear after only %d frames", i+1)
		}
	}
	if got := analyzer.Observe(clear); got != Clear {
		t.Fatalf("state after stable frames = %v, want Clear", got)
	}
}

func TestStaleDepthForcesUnknown(t *testing.T) {
	var analyzer Analyzer
	analyzer.Observe(Observation{NearestMM: 2000, Coverage: 60, Tracking: true})
	analyzer.MarkUnavailable()
	if got := analyzer.State(); got != Unknown {
		t.Fatalf("state after stale deadline = %v, want Unknown", got)
	}
}
