package draw

import (
	"testing"
	"time"
)

func TestClockAndNote(t *testing.T) {
	t.Parallel()

	if !Clock(false).IsZero() {
		t.Fatal("off")
	}
	if Clock(true).IsZero() {
		t.Fatal("on")
	}

	var d time.Duration
	Note(false, &d, time.Now())
	if d != 0 {
		t.Fatal("untracked")
	}
	Note(true, &d, time.Now().Add(-time.Millisecond))
	if d <= 0 {
		t.Fatal("tracked")
	}
}
