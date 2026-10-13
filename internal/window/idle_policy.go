package window

import "time"

// paceAction is what one drawn frame asks the pacer to do next.
type paceAction int

const (
	// paceHold leaves the mode alone.
	paceHold paceAction = iota
	// paceToIdle switches to the minimum FPS mode.
	paceToIdle
	// paceToFast switches back to vsync.
	paceToFast
	// paceBurst stays idle and schedules one frame immediately, so a
	// changing page is stepped again before the keep-alive interval.
	paceBurst
)

// paceCleanRun is the clean-frame streak that lets a window idle. It is
// shorter than the input grace on purpose: a page that paints one small
// rect every sixth frame, like a progress bar on a 10 Hz cadence, would
// otherwise pin the window to 60 frames per second forever.
const paceCleanRun = 5

// idleBurst is the consecutive dirty idle frames that count as an animation
// rather than one painted change.
const idleBurst = 2

// pacePolicy defaults. The keep-alive is the idle tick rate: page ticks
// still run there and can detect changes.
const (
	paceGrace     = 250 * time.Millisecond
	paceKeepAlive = 125 * time.Millisecond
)

// pacePolicy decides when a window may leave the vsync loop. It is pure
// state: callers pass the time, so tests need no window. In fast mode it
// idles after paceCleanRun clean frames and paceGrace without input; idle
// mode returns to vsync at once on input, and after idleBurst dirty frames.
type pacePolicy struct {
	grace     time.Duration
	keepAlive time.Duration
	fast      bool
	lastInput time.Time
	pending   time.Time
	cleanRun  int
	dirtyRun  int
}

func newPacePolicy(now time.Time, grace, keepAlive time.Duration) pacePolicy {
	return pacePolicy{
		grace:     grace,
		keepAlive: keepAlive,
		fast:      true,
		pending:   now,
	}
}
