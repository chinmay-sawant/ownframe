package window

import "time"

// paceStep is one frame at the 60 Hz the fast loop runs, the unit the skipped
// counter uses.
const paceStep = 16 * time.Millisecond

const keepAliveHalfDiv = 2

// paceWantState is the page-driven half of the pacing policy: the deadline
// the page asked for, the previous step time, and the counters the devtools
// Frame tab shows. It is embedded so idle_policy.go stays inside the file
// size limit.
type paceWantState struct {
	want     time.Time
	last     time.Time
	skipped  uint64
	presents uint64
}

// next reports the deadline the pacer is working to: the page deadline while
// one is pending, the keep-alive schedule when idle without one, and the zero
// time on the fast path when the page did not ask. The devtools Frame tab
// reads it, and wait reads it to size the timer sleep.
func (p *pacePolicy) next() time.Time {
	if !p.want.IsZero() {
		return p.want
	}

	if p.fast {
		return time.Time{}
	}

	return p.pending.Add(p.keepAlive)
}

// wait reports how long the timer may sleep before its next check. It runs on
// the timer goroutine under the same lock as step and due.
func (p *pacePolicy) wait(now time.Time) time.Duration {
	if p.fast {
		return p.keepAlive / keepAliveHalfDiv
	}

	if d := p.next().Sub(now); d > time.Millisecond {
		return d
	}

	return time.Millisecond
}

// skip counts the vsync frames the gap since the previous step covers, minus
// the frame the step presents, so a 60 Hz loop reports none and a one second
// sleep reports about 61. A policy that was not built by newPacePolicy has no
// previous time and records nothing.
func (p *pacePolicy) skip(now time.Time) {
	if p.last.IsZero() {
		p.last = now

		return
	}

	gap := now.Sub(p.last)
	p.last = now

	if gap <= paceStep {
		return
	}

	steps := gap / paceStep
	if steps < 1 {
		return
	}

	p.skipped += uint64(steps) - 1
}
