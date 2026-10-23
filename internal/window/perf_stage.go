package window

import "time"

// perfStage holds per-frame section timings.
type perfStage struct {
	update        time.Duration
	tick          time.Duration
	sync          time.Duration
	draw          time.Duration
	present       time.Duration
	pendingUpdate time.Duration
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
	d := s.dev.stages.pendingUpdate + s.dev.stages.draw + s.dev.stages.present
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
