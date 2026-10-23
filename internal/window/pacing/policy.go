// Package pacing decides frame deadlines and idle transitions without a graphics host.
// Callers supply time and serialize access to each Policy.
package pacing

import "time"

// Action is what one drawn frame asks the pacer to do next.
type Action int

const (
	// Hold leaves the mode alone.
	Hold Action = iota
	// ToIdle switches to the minimum FPS mode.
	ToIdle
	// ToFast switches back to vsync.
	ToFast
	// Burst stays idle and schedules one frame immediately, so a
	// changing page is stepped again before the keep-alive interval.
	Burst
)

// cleanFrames is the clean-frame streak that lets a window idle. It is
// shorter than the input grace on purpose: a page that paints one small
// rect every sixth frame, like a progress bar on a 10 Hz cadence, would
// otherwise pin the window to 60 frames per second forever.
const cleanFrames = 5

// idleBurst is the consecutive dirty idle frames that count as an animation
// rather than one painted change.
const idleBurst = 2

// Policy defaults. The keep-alive is the idle tick rate: page ticks
// still run there and can detect changes.
const (
	paceGrace     = 250 * time.Millisecond
	paceKeepAlive = 125 * time.Millisecond
)

// Policy decides when a window may leave the vsync loop. It is pure
// state: callers pass the time, so tests need no window. In fast mode it
// idles after cleanFrames clean frames and paceGrace without input; idle
// mode returns to vsync at once on input, on an immediate page request, and
// after idleBurst dirty frames that came without a declared deadline. A page
// deadline still ahead schedules the next idle wake; with no deadline the
// keep-alive fallback does.
type Policy struct {
	fast      bool
	lastInput time.Time
	pending   time.Time
	cleanRun  int
	dirtyRun  int
	want      time.Time
	last      time.Time
	skipped   uint64
	presents  uint64
}

// New starts a fast policy with a 250 ms input grace and 125 ms idle keep-alive.
func New(now time.Time) Policy {
	return Policy{
		fast:    true,
		pending: now,
	}
}

// Step records one drawn frame at now. dirty reports that the frame painted
// and input that keyboard, pointer, wheel, or touch activity arrived since
// the previous step. want is the page's next-frame deadline for this frame,
// read once per tick; the zero time means the page did not ask and the
// keep-alive fallback schedules the wake. A want at or before now asks for
// the next frame at once, so a page that asks that way every frame never
// leaves the fast loop.
func (p *Policy) Step(now time.Time, dirty, input bool, want time.Time) Action {
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
			return Hold
		}

		if p.cleanRun < cleanFrames || now.Sub(p.lastInput) < paceGrace {
			return Hold
		}

		p.fast = false
		p.dirtyRun = 0
		p.pending = now

		return ToIdle
	}

	if input {
		p.fast = true

		return ToFast
	}

	if !want.IsZero() && !want.After(now) {
		p.fast = true

		return ToFast
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

				return ToFast
			}

			p.pending = now

			return Burst
		}

		p.pending = now

		return Hold
	}

	p.dirtyRun = 0

	return Hold
}

// Due reports whether the schedule should post a frame now, and consumes the
// slot. The schedule only runs while idle. A page deadline that has arrived
// posts at once; while one is still pending the keep-alive fallback does not
// run, so a page that asked for a one second frame is not woken every
// keep-alive. With no page deadline the fallback posts every keep-alive
// interval, so a page that never asks cannot stall.
func (p *Policy) Due(now time.Time) bool {
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

	if now.Sub(p.pending) < paceKeepAlive {
		return false
	}

	p.pending = now

	return true
}

// frameStep is one frame at the 60 Hz the fast loop runs, the unit the skipped
// counter uses.
const frameStep = 16 * time.Millisecond

const keepAliveHalfDiv = 2

// next reports the deadline the pacer is working to: the page deadline while
// one is pending, the keep-alive schedule when idle without one, and the zero
// time on the fast path when the page did not ask. The devtools Frame tab
// reads it, and Wait reads it to size the timer sleep.
func (p *Policy) next() time.Time {
	if !p.want.IsZero() {
		return p.want
	}

	if p.fast {
		return time.Time{}
	}

	return p.pending.Add(paceKeepAlive)
}

// Wait reports how long the timer may sleep before its next check.
// The host must serialize it with Step and Due.
func (p *Policy) Wait(now time.Time) time.Duration {
	if p.fast {
		return paceKeepAlive / keepAliveHalfDiv
	}

	if d := p.next().Sub(now); d > time.Millisecond {
		return d
	}

	return time.Millisecond
}

// skip counts the vsync frames the gap since the previous step covers, minus
// the frame the step presents, so a 60 Hz loop reports none and a one second
// sleep reports about 61. A policy that was not built by New has no
// previous time and records nothing.
func (p *Policy) skip(now time.Time) {
	if p.last.IsZero() {
		p.last = now

		return
	}

	gap := now.Sub(p.last)
	p.last = now

	if gap <= frameStep {
		return
	}

	steps := gap / frameStep
	if steps < 1 {
		return
	}

	p.skipped += uint64(steps) - 1
}

// Snapshot contains the current schedule and cumulative diagnostic counters.
type Snapshot struct {
	Next     time.Time
	Skipped  uint64
	Presents uint64
	Idle     bool
}

// Snapshot returns a copy. Reading it does not consume a deadline.
func (p *Policy) Snapshot() Snapshot {
	return Snapshot{Next: p.next(), Skipped: p.skipped, Presents: p.presents, Idle: !p.fast}
}
