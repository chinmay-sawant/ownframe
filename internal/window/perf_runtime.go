package window

import (
	"runtime"
	"time"
)

// perfSampleGap bounds MemStats cost to about twice a second.
const (
	perfSampleGap     = 500 * time.Millisecond
	perfIdleSampleGap = 2 * time.Second
)

// perfSampler holds the last sampled runtime numbers.
type perfSampler struct {
	last           time.Time
	first          bool
	total          uint64
	frames         uint64
	heap           uint64
	alloc          uint64
	rss            uint64
	goroutines     int
	gap            time.Duration
	heapSys        uint64
	heapInuse      uint64
	heapIdle       uint64
	heapReleased   uint64
	sys            uint64
	stackSys       uint64
	pauseTotal     uint64
	numGC          uint32
	allocPerSec    uint64
	repaints       uint64
	repaintsPerSec float64
}

// maybe refreshes the sample when the gap has passed.
func (p *perfSampler) maybe(now time.Time) bool {
	gap := p.gap
	if gap == 0 {
		gap = perfSampleGap
	}
	if !p.last.IsZero() && now.Sub(p.last) < gap {
		return false
	}
	elapsed := now.Sub(p.last)
	p.last = now
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if !p.first {
		p.first = true
		p.alloc = 0
	} else {
		frames := p.frames
		if frames == 0 {
			frames = 1
		}
		p.alloc = (m.TotalAlloc - p.total) / frames
		p.allocPerSec = uint64(float64(m.TotalAlloc-p.total) / elapsed.Seconds())
	}
	p.frames = 0
	p.total = m.TotalAlloc
	p.heap = m.HeapAlloc
	p.heapSys = m.HeapSys
	p.heapInuse = m.HeapInuse
	p.heapIdle = m.HeapIdle
	p.heapReleased = m.HeapReleased
	p.sys = m.Sys
	p.stackSys = m.StackSys
	p.pauseTotal = m.PauseTotalNs
	p.numGC = m.NumGC
	p.goroutines = runtime.NumGoroutine()
	p.rss = perfRSS()
	return true
}

// perfSampleRuntime gates the sampler and publishes into devState. It is a
// no-op unless the shell opted into Perf.
func (s *shell) perfSampleRuntime() {
	if !s.perf {
		return
	}
	p := &s.dev.sampler
	p.gap = perfSampleGap
	if s.pace.idle {
		p.gap = perfIdleSampleGap
	}
	previous := p.last
	if p.maybe(time.Now()) {
		if insp := s.devInspector(); insp != nil {
			repaints := insp.Stats().Repaints
			if !previous.IsZero() && repaints >= p.repaints {
				p.repaintsPerSec = float64(repaints-p.repaints) / p.last.Sub(previous).Seconds()
			}
			p.repaints = repaints
		}
		s.dev.runHeap = p.heap
		s.dev.runAlloc = p.alloc
		s.dev.runRSS = p.rss
		s.dev.runGoroutines = p.goroutines
	}
}

func (s *shell) runtimeSnapshot() map[string]any {
	s.perfSampleRuntime()
	p := &s.dev.sampler
	return map[string]any{
		"heap_alloc": p.heap, "heap_sys": p.heapSys, "total_alloc": p.total,
		"num_gc": p.numGC, "goroutines": p.goroutines, "rss_bytes": p.rss,
		"sampled_at": p.last, "alloc_per_sec": p.allocPerSec,
		"heap_inuse": p.heapInuse, "heap_idle": p.heapIdle,
		"heap_released": p.heapReleased, "sys": p.sys,
		"stack_sys": p.stackSys, "gc_pause_total_ns": p.pauseTotal,
	}
}

// perfRuntime returns the last published runtime numbers. A zero RSS means
// the platform did not report one; genuine resident memory is never zero,
// so the devtools row renders that case as "-" rather than "0B".
func (s *shell) perfRuntime() (heap, rss, alloc uint64, goroutines int) {
	return s.dev.runHeap, s.dev.runRSS, s.dev.runAlloc, s.dev.runGoroutines
}
