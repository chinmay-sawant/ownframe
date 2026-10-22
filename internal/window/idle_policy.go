package window

import (
	"time"
)

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
// mode returns to vsync at once on input, on an immediate page request, and
// after idleBurst dirty frames that came without a declared deadline. A page
// deadline still ahead schedules the next idle wake; with no deadline the
// keep-alive fallback does.
type pacePolicy struct {
	grace     time.Duration
	keepAlive time.Duration
	fast      bool
	lastInput time.Time
	pending   time.Time
	cleanRun  int
	dirtyRun  int
	paceWantState
}

func newPacePolicy(now time.Time, grace, keepAlive time.Duration) pacePolicy {
	return pacePolicy{
		grace:     grace,
		keepAlive: keepAlive,
		fast:      true,
		pending:   now,
	}
}

// step records one drawn frame at now. dirty reports that the frame painted
// and input that keyboard, pointer, wheel, or touch activity arrived since
// the previous step. want is the page's next-frame deadline for this frame,
// read once per tick; the zero time means the page did not ask and the
// keep-alive fallback schedules the wake. A want at or before now asks for
// the next frame at once, so a page that asks that way every frame never
// leaves the fast loop.
func (p *pacePolicy) step(now time.Time, dirty, input bool, want time.Time) paceAction {
	if input {
		p.lastInput = now
	}

	p.skip(now)
	p.want = want
	p.presents++

	if dirty {
		p.cleanRun = 0
	} else {
		p.cleanRun++
	}

	if p.fast {
		if !want.IsZero() && !want.After(now) {
			return paceHold
		}

		if p.cleanRun < paceCleanRun || now.Sub(p.lastInput) < p.grace {
			return paceHold
		}

		p.fast = false
		p.dirtyRun = 0
		p.pending = now

		return paceToIdle
	}

	if input {
		p.fast = true

		return paceToFast
	}

	if !want.IsZero() && !want.After(now) {
		p.fast = true

		return paceToFast
	}

	if dirty {
		p.dirtyRun++

		// A dirty frame with no declared deadline keeps the old escalation:
		// two in a row read as an animation. A page that declared its next
		// frame owns the schedule, so one dirty frame stays idle and the
		// timer wakes at the declared time.
		if p.want.IsZero() {
			if p.dirtyRun >= idleBurst {
				p.fast = true

				return paceToFast
			}

			p.pending = now

			return paceBurst
		}

		p.pending = now

		return paceHold
	}

	p.dirtyRun = 0

	return paceHold
}

// due reports whether the schedule should post a frame now, and consumes the
// slot. The schedule only runs while idle. A page deadline that has arrived
// posts at once; while one is still pending the keep-alive fallback does not
// run, so a page that asked for a one second frame is not woken every
// keep-alive. With no page deadline the fallback posts every keep-alive
// interval, so a page that never asks cannot stall.
func (p *pacePolicy) due(now time.Time) bool {
	if p.fast {
		return false
	}

	if !p.want.IsZero() {
		if p.want.After(now) {
			return false
		}

		p.want = time.Time{}
		p.pending = now

		return true
	}

	if now.Sub(p.pending) < p.keepAlive {
		return false
	}

	p.pending = now

	return true
}
