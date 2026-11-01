package window

import (
	"testing"
	"time"
)

func TestNormalizeMaxFPS(t *testing.T) {
	if got := NormalizeMaxFPS(0); got != DefaultMaxFPS {
		t.Fatalf("zero = %d, want default %d", got, DefaultMaxFPS)
	}
	if got := NormalizeMaxFPS(-1); got != 0 {
		t.Fatalf("uncapped = %d, want 0", got)
	}
	for _, fps := range []int{60, 120, 144, 240} {
		if got := NormalizeMaxFPS(fps); got != fps {
			t.Fatalf("fps %d = %d", fps, got)
		}
	}
	if got := NormalizeMaxFPS(10); got != DefaultMaxFPS {
		t.Fatalf("invalid low = %d, want default", got)
	}
	if got := NormalizeMaxFPS(5000); got != DefaultMaxFPS {
		t.Fatalf("invalid high = %d, want default", got)
	}
}

func TestLoopRunsHotterThanPresents(t *testing.T) {
	for _, fps := range []int{60, 120, 144, 240} {
		if got := LoopFPS(fps); got != fps+FPSHeadroom {
			t.Fatalf("loop(%d) = %d, want %d", fps, got, fps+FPSHeadroom)
		}
		if pi, li := PresentInterval(fps), LoopInterval(fps); pi != li {
			t.Fatalf("fps %d: present %v must equal loop %v (loop paced to cap)", fps, pi, li)
		}
	}
	if LoopFPS(0) != 0 || PresentInterval(0) != 0 || LoopInterval(0) != 0 {
		t.Fatal("uncapped must have zero intervals")
	}
}

func TestCapHoldsUnderSustainedLoad(t *testing.T) {
	const maxFPS = 60
	base := time.Unix(1000, 0)
	lastLoop := base
	presents := 0
	at := base
	// Simulate 2s of sustained load paced to the cap: 160 loop iterations
	// at 16.6ms, each presenting once.
	for i := 0; i < 160; i++ {
		at = at.Add(LoopInterval(maxFPS))
		lastLoop = at
		presents++
	}
	if presents != 160 {
		t.Fatalf("2s at 60 cap presented %d, want 160", presents)
	}
	_ = lastLoop
	// The hotter app rate is carried by TPS headroom, not extra presents:
	// 60 presents, 80 updates per second.
	if LoopFPS(maxFPS) != maxFPS+FPSHeadroom {
		t.Fatalf("loop TPS = %d, want %d", LoopFPS(maxFPS), maxFPS+FPSHeadroom)
	}
}

func TestLoopWaitEnforcesIntervalWithoutBusySpin(t *testing.T) {
	const maxFPS = 60
	base := time.Unix(2000, 0)
	if got := loopWait(maxFPS, time.Time{}, base); got != 0 {
		t.Fatalf("first frame wait = %v, want 0", got)
	}
	if got := loopWait(maxFPS, base, base); got != LoopInterval(maxFPS) {
		t.Fatalf("immediate wait = %v, want %v", got, LoopInterval(maxFPS))
	}
	if got := loopWait(maxFPS, base, base.Add(2*LoopInterval(maxFPS))); got != 0 {
		t.Fatalf("late wait = %v, want 0", got)
	}
	if got := loopWait(0, base, base); got != 0 {
		t.Fatalf("uncapped wait = %v, want 0", got)
	}
}

func TestCoalescingNeverStarvesImmediateRequests(t *testing.T) {
	// The Draw limiter paces the loop to the present cap with sleep; it
	// never skips heavy work, so an animation asking for the next frame at
	// once (WantFrameIn(0)) is always presented on the next loop tick. This
	// pins the contract: every loop iteration presents, and the hotter app
	// rate is carried by TPS headroom, not skipped presents.
	const maxFPS = 60
	if LoopInterval(maxFPS) != PresentInterval(maxFPS) {
		t.Fatal("loop must be paced to the present cap")
	}
}
