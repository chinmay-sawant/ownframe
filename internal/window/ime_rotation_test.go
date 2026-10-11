package window

import (
	"testing"
	"time"
)

func TestIMERotationKeepsFocusAcrossKeyboardHide(t *testing.T) {
	start := time.Unix(1, 0)
	var r imeRotationState
	r.layout(420, 900, start)
	if r.endByUser(start.Add(time.Millisecond)) {
		t.Fatal("end should wait until resize confirms rotation")
	}
	if wait, keep := r.resolve(start.Add(100 * time.Millisecond)); !wait || keep {
		t.Fatalf("before resize wait=%v keep=%v", wait, keep)
	}
	r.layout(900, 420, start.Add(110*time.Millisecond))
	if wait, keep := r.resolve(start.Add(120 * time.Millisecond)); wait || !keep {
		t.Fatalf("after resize wait=%v keep=%v", wait, keep)
	}
}

func TestIMEUserDismissStillEndsAfterNoResize(t *testing.T) {
	start := time.Unix(1, 0)
	var r imeRotationState
	r.layout(420, 900, start)
	if r.endByUser(start.Add(time.Millisecond)) {
		t.Fatal("unexpected rotation")
	}
	if wait, keep := r.resolve(start.Add(100 * time.Millisecond)); !wait || keep {
		t.Fatalf("early end wait=%v keep=%v", wait, keep)
	}
	if wait, keep := r.resolve(start.Add(300 * time.Millisecond)); wait || keep {
		t.Fatalf("user dismissal wait=%v keep=%v", wait, keep)
	}
}

func TestIMEHideAfterRotationKeepsFocus(t *testing.T) {
	start := time.Unix(1, 0)
	var r imeRotationState
	r.layout(420, 900, start)
	r.layout(900, 420, start.Add(100*time.Millisecond))
	if !r.endByUser(start.Add(150 * time.Millisecond)) {
		t.Fatal("rotation hide should retain focus")
	}
}
