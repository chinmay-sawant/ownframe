package window

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"runtime"
	"slices"
	"time"
)

// perfSource is a screen that reports its own opt-in. The page implements
// it; a custom screen without it relies on DevTools or Options.Perf.
type perfSource interface{ Perf() bool }

// perfEnabled reports whether the screen opted into sampling.
func perfEnabled(app host.Screen) bool {
	src, ok := app.(perfSource)

	return ok && src.Perf()
}

// NewGameWithOptions returns the screen loop with perf sampling configured.
// It keeps the NewGame behavior for every other setting.
func NewGameWithOptions(ctx context.Context, app host.Screen, options Options) ebiten.Game {
	game := newShell(ctx, app)
	game.perfOption = options.Perf
	game.perf = options.Perf || perfEnabled(app)
	game.wirePerf()

	return game
}

// wirePerf installs this window's opt-in performance readers.
func (s *shell) wirePerf() {
	s.perfHooks.devFramePerf = func() (avg, p95, p99 time.Duration, long uint64, ok bool) {
		if !s.perf {
			return 0, 0, 0, 0, false
		}
		a, p95v, p99v, l := s.perfSummary()
		if l < 0 {
			l = 0
		}

		return a, p95v, p99v, uint64(l), true
	}
	s.perfHooks.devPipelinePerf = func(st host.Stats) (tpl, lay, dsp, pnt time.Duration, dOps, dReg, chOps int, ok bool) {
		if !s.perf {
			return 0, 0, 0, 0, 0, 0, 0, false
		}

		return st.LastTemplate, st.StyleTime, st.LayoutTime, st.PaintTime, st.DirtyOps, st.DirtyRegions, st.ChangedOps, true
	}
	s.perfHooks.devRuntimePerf = func() (alloc, heap, rss uint64, gr int, ok bool) {
		if !s.perf {
			return 0, 0, 0, 0, false
		}
		h, r, a, g := s.perfRuntime()

		return a, h, r, g, true
	}
}

const perfRingCap = 120

const perfLongThreshold = time.Second / 60

const perfP95Quantile = 0.95

const perfP99Quantile = 0.99

const perfMinSamples = 2

// perfRing keeps the last perfRingCap frame durations. order is scratch for
// summaries; it lives inside the ring so a summary allocates nothing.
type perfRing struct {
	buf   [perfRingCap]time.Duration
	order [perfRingCap]time.Duration
	n     int
	next  int
}

func (r *perfRing) record(d time.Duration) {
	if d < 0 {
		d = 0
	}
	r.buf[r.next] = d
	r.next = (r.next + 1) % perfRingCap
	if r.n < perfRingCap {
		r.n++
	}
}

func (r *perfRing) count() int { return r.n }

// sorted copies the ring in age order and sorts it ascending. It reuses
// r.order, so repeated summaries allocate nothing.
func (r *perfRing) sorted() []time.Duration {
	s := r.order[:r.n]
	old := (r.next - r.n + perfRingCap) % perfRingCap
	for i := range s {
		s[i] = r.buf[(old+i)%perfRingCap]
	}
	slices.Sort(s)

	return s
}

// summary refreshes every cached reading in one walk of the ring.
func (r *perfRing) summary() (avg, p95, p99 time.Duration, long int) {
	if r.n == 0 {
		return 0, 0, 0, 0
	}

	s := r.sorted()
	var sum time.Duration

	for _, d := range s {
		sum += d
		if d > perfLongThreshold {
			long++
		}
	}

	return sum / time.Duration(r.n), quantile(s, perfP95Quantile), quantile(s, perfP99Quantile), long
}

// avg, p95, p99, and longFrames are the single-reading views of summary.
func (r *perfRing) avg() time.Duration {
	avg, _, _, _ := r.summary()

	return avg
}

func (r *perfRing) p95() time.Duration {
	_, p95, _, _ := r.summary()

	return p95
}

func (r *perfRing) p99() time.Duration {
	_, _, p99, _ := r.summary()

	return p99
}

func (r *perfRing) longFrames(th time.Duration) int {
	if th != perfLongThreshold {
		c := 0
		for _, d := range r.sorted() {
			if d > th {
				c++
			}
		}

		return c
	}

	_, _, _, long := r.summary()

	return long
}

// quantile interpolates the q quantile from an ascending slice.
func quantile(s []time.Duration, q float64) time.Duration {
	n := len(s)
	if n == 0 {
		return 0
	}

	at := q * float64(n-1)
	lo := int(at)
	if lo >= n-1 {
		return s[n-1]
	}

	fr := at - float64(lo)

	return s[lo] + time.Duration(fr*float64(s[lo+1]-s[lo]))
}

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

// perfStage holds per-frame section timings.
type perfStage struct {
	update        time.Duration
	tick          time.Duration
	sync          time.Duration
	draw          time.Duration
	present       time.Duration
	pendingUpdate time.Duration
	work          time.Duration
}

// perfTimed times f into d.
func perfTimed(d *time.Duration, f func()) {
	start := time.Now()
	f()
	*d = time.Since(start)
}

// perfTimedErr times f into d and returns its error.
func perfTimedErr(d *time.Duration, f func() error) error {
	start := time.Now()
	err := f()
	*d = time.Since(start)
	return err
}

// perfFrameSample records Update and Draw work, excluding time between calls, into
// the ring and refreshes the
// cached summary. It is a no-op unless the shell opted into Perf.
func (s *shell) perfFrameSample() {
	if !s.perf {
		return
	}
	s.dev.sampler.frames++
	d := s.dev.stages.pendingUpdate + s.dev.stages.draw + s.dev.stages.present
	s.dev.stages.work = d
	s.dev.stages.pendingUpdate = 0
	if d > 0 {
		s.dev.ring.record(d)
		avg, p95, p99, long := s.dev.ring.summary()
		s.dev.perfAvg, s.dev.perfP95, s.dev.perfP99, s.dev.perfLong = avg, p95, p99, long
	}
	s.perfSampleRuntime()
}

// timedErr times f into d with Perf on and runs it plainly with Perf off,
// so an end-user window pays no clock calls on its hot path.
func (s *shell) timedErr(d *time.Duration, f func() error) error {
	if !s.perf {
		return f()
	}

	return perfTimedErr(d, f)
}

// perfSummary returns the cached frame stats; zero on an empty ring.
func (s *shell) perfSummary() (avg, p95, p99 time.Duration, long int) {
	return s.dev.perfAvg, s.dev.perfP95, s.dev.perfP99, s.dev.perfLong
}
