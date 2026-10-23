package page

import "time"

// allocSampleGap bounds ReadMemStats on a Perf redraw to about twice a second.
// Stage timing still records every redraw.
const allocSampleGap = 500 * time.Millisecond

// SetPerf turns pipeline timing, dirty counters, and alloc tracking on or
// off. It is off by default, so a shipped app pays nothing; a developer
// opts in to fill the DevTools performance rows and the matching Stats.
// Each call clears the allocator sample clock, so the next tracked redraw
// samples immediately.
func (p *Page) SetPerf(on bool) {
	p.perf = on
	p.allocAt = time.Time{}
}

// allocDue reports whether this redraw should read the allocator.
// A zero clock is due. A later call inside allocSampleGap is not and does
// not move the clock. A call at the gap is due.
func (p *Page) allocDue(now time.Time) bool {
	if !p.allocAt.IsZero() && now.Sub(p.allocAt) < allocSampleGap {
		return false
	}
	p.allocAt = now

	return true
}

// Perf reports whether pipeline timing and dirty counters are recorded.
func (p *Page) Perf() bool { return p.perf }

// clearPerf zeroes the opt-in Stats fields after an untracked Redraw, so a
// page that opts out never shows stale developer numbers.
func (p *Page) clearPerf() {
	p.stats.lastTemplate = 0
	p.stats.styleTime = 0
	p.stats.layoutTime = 0
	p.stats.paintTime = 0
	p.stats.otherTime = 0
	p.stats.dirtyOps = 0
	p.stats.dirtyRegions = 0
	p.stats.dirtyArea = 0
	p.stats.changedOps = 0
	p.stats.allocFrame = 0
}
