package window

import "time"

// due reports whether the keep-alive schedule should post a frame now, and
// consumes the interval. The schedule only runs while idle.
func (p *pacePolicy) due(now time.Time) bool {
	if p.fast || now.Sub(p.pending) < p.keepAlive {
		return false
	}

	p.pending = now

	return true
}
