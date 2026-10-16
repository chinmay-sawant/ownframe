package page

import (
	"runtime"
	"time"
)

// beginRedraw samples allocs when Perf is on and clears stale numbers
// when off, then counts the redraw. finishRedraw records the frame.
func (p *Page) beginRedraw() (bool, runtime.MemStats, time.Time) {
	track := p.perf
	var m0 runtime.MemStats
	if track {
		m0 = allocStart()
	} else {
		p.clearPerf()
	}
	start := time.Now()
	p.stats.redraws++
	return track, m0, start
}

func (p *Page) finishRedraw(track bool, m0 runtime.MemStats, start time.Time) {
	p.stats.lastRedraw = time.Since(start)
	if track {
		p.stats.allocFrame = allocUsed(m0)
	}
}

// noteTime records one stage duration only when Perf is on, so the
// stages read clean and an untracked Redraw pays no clock beyond start.
func (p *Page) noteTime(track bool, dst *time.Duration, start time.Time) {
	if track {
		*dst = time.Since(start)
	}
}
