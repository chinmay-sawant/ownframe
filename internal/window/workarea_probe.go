package window

import "time"

const workareaSettleInterval = 100 * time.Millisecond

// workareaProbe waits for the native maximize configure to settle. The
// maximized flag can arrive before the position and size events.
type workareaProbe struct {
	area  workarea
	since time.Time
}

func (p *workareaProbe) observe(area workarea, now time.Time) bool {
	if p.since.IsZero() || p.area != area {
		p.area, p.since = area, now
		return false
	}
	return now.Sub(p.since) >= workareaSettleInterval
}
