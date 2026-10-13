package window

import "time"

// step records one drawn frame at now. dirty reports that the frame painted
// and input that keyboard, pointer, wheel, or touch activity arrived since
// the previous step.
func (p *pacePolicy) step(now time.Time, dirty, input bool) paceAction {
	if input {
		p.lastInput = now
	}

	if dirty {
		p.cleanRun = 0
	} else {
		p.cleanRun++
	}

	if p.fast {
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

	if dirty {
		p.dirtyRun++
		if p.dirtyRun >= idleBurst {
			p.fast = true

			return paceToFast
		}

		p.pending = now

		return paceBurst
	}

	p.dirtyRun = 0

	return paceHold
}
