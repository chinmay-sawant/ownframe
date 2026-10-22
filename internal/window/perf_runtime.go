package window

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// perfSampleGap bounds MemStats cost to about twice a second.
const perfSampleGap = 500 * time.Millisecond

// perfSampler holds the last sampled runtime numbers.
type perfSampler struct {
	last       time.Time
	first      bool
	total      uint64
	frames     uint64
	heap       uint64
	alloc      uint64
	rss        uint64
	goroutines int
}

// maybe refreshes the sample when the gap has passed.
func (p *perfSampler) maybe(now time.Time) bool {
	p.frames++
	if !p.last.IsZero() && now.Sub(p.last) < perfSampleGap {
		return false
	}
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
	}
	p.frames = 0
	p.total = m.TotalAlloc
	p.heap = m.HeapAlloc
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
	if s.dev.sampler.maybe(time.Now()) {
		p := &s.dev.sampler
		s.dev.runHeap = p.heap
		s.dev.runAlloc = p.alloc
		s.dev.runRSS = p.rss
		s.dev.runGoroutines = p.goroutines
	}
}

// perfRuntime returns the last published runtime numbers.
func (s *shell) perfRuntime() (heap, rss, alloc uint64, goroutines int) {
	return s.dev.runHeap, s.dev.runRSS, s.dev.runAlloc, s.dev.runGoroutines
}

// perfRSS reads resident bytes on linux without cgo; 0 elsewhere.
func perfRSS() uint64 {
	if runtime.GOOS != "linux" {
		return 0
	}
	b, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}
	f := strings.Fields(string(b))
	if len(f) < perfMinSamples {
		return 0
	}
	pages, err := strconv.ParseUint(f[1], 10, 64)
	if err != nil {
		return 0
	}
	return pages * uint64(max(os.Getpagesize(), 0))
}
