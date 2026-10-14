package window

import "time"

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
