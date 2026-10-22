package page

import (
	"runtime"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
)

// beginRedraw samples allocs when Perf is on and clears stale numbers
// when off, then counts the redraw. finishRedraw records the frame.
func (p *Page) beginRedraw() (bool, runtime.MemStats, time.Time) {
	track := p.perf
	var m0 runtime.MemStats
	if track {
		m0 = draw.AllocStart()
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
		p.stats.allocFrame = draw.AllocUsed(m0)
	}
}
