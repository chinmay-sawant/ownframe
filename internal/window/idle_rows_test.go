package window

import (
	"strings"
	"testing"
	"time"
)

// TestDevFramePacingRows pins the scheduler state in the Frame tab.
func TestDevFramePacingRows(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.pace.idle = true
	s.pace.skipped = 41
	s.pace.next = time.Now().Add(750 * time.Millisecond)

	rows := s.devFrameRows(320)

	if got := devFrameValue(rows, "Idle"); got != "true" {
		t.Fatalf("Idle row = %q, want true", got)
	}
	if got := devFrameValue(rows, "Pace skip"); got != "41" {
		t.Fatalf("Pace skip row = %q, want 41", got)
	}
	if got := devFrameValue(rows, "Next"); !strings.HasSuffix(got, "ms") {
		t.Fatalf("Next row = %q, want a millisecond value", got)
	}
}

// TestDevPaceNextEmpty pins the fast-path rendering.
func TestDevPaceNextEmpty(t *testing.T) {
	if got := devPaceNext(time.Time{}); got != "-" {
		t.Fatalf("devPaceNext(zero) = %q, want -", got)
	}
}
