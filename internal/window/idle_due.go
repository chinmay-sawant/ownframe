package window

import "time"

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
