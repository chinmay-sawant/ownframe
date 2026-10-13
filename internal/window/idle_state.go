package window

import (
	"fmt"
	"time"
)

// paceState is the pacer's published state for the devtools Frame tab: the
// next scheduled idle wake, the vsync frames skipped so far, and whether the
// window is parked. The wrapper writes it on the game goroutine after each
// frame; the tab reads it there. It lives in its own file so game_state.go
// stays inside the file size limit.
type paceState struct {
	next    time.Time
	skipped uint64
	idle    bool
}

// devPaceNext renders the time until the next scheduled idle wake, or "-"
// when the pacer is on the vsync loop and nothing is scheduled.
func devPaceNext(next time.Time) string {
	if next.IsZero() {
		return "-"
	}

	if d := time.Until(next); d > 0 {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	return "0ms"
}
