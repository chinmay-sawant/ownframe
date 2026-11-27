// Package monitor holds Quarry's UI-free monitoring state: bounded metric
// rings, rate and latency math, slow-query capture, and the background
// collector whose results the UI drains from SetTick.
package monitor

import (
	"slices"
	"sort"
	"sync"
	"time"
)

// DefaultCapacity is the default number of samples a MetricRing keeps.
const DefaultCapacity = 120

// MetricRing is a fixed-capacity ring buffer: pushing past capacity overwrites
// the oldest sample, so memory never grows. It is not safe for concurrent use;
// the UI owns it and fills it from SetTick.
type MetricRing[T any] struct {
	buf   []T
	start int // index of the oldest sample
	n     int
}

// NewRing returns a ring holding up to capacity samples (DefaultCapacity when
// capacity <= 0).
func NewRing[T any](capacity int) *MetricRing[T] {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &MetricRing[T]{buf: make([]T, capacity)}
}

// Cap returns the maximum number of samples.
func (r *MetricRing[T]) Cap() int { return len(r.buf) }

// Len returns the number of samples currently held.
func (r *MetricRing[T]) Len() int { return r.n }

// Push appends v, evicting the oldest sample when full.
func (r *MetricRing[T]) Push(v T) {
	if r.n < len(r.buf) {
		r.buf[(r.start+r.n)%len(r.buf)] = v
		r.n++
		return
	}
	r.buf[r.start] = v
	r.start = (r.start + 1) % len(r.buf)
}

// At returns the i-th sample, 0 being the oldest.
func (r *MetricRing[T]) At(i int) T {
	if i < 0 || i >= r.n {
		panic("monitor: ring index out of range")
	}
	return r.buf[(r.start+i)%len(r.buf)]
}

// Last returns the newest sample.
func (r *MetricRing[T]) Last() (T, bool) {
	var zero T
	if r.n == 0 {
		return zero, false
	}
	return r.At(r.n - 1), true
}

// Snapshot copies the samples, oldest first.
func (r *MetricRing[T]) Snapshot() []T {
	out := make([]T, r.n)
	for i := range out {
		out[i] = r.At(i)
	}
	return out
}

// Reset empties the ring without reallocating.
func (r *MetricRing[T]) Reset() {
	clear(r.buf)
	r.start, r.n = 0, 0
}

// Series maps every sample (oldest first) to a float for charting.
func Series[T any](r *MetricRing[T], f func(T) float64) []float64 {
	out := make([]float64, r.Len())
	for i := range out {
		out[i] = f(r.At(i))
	}
	return out
}

// Range returns the min and max of vals; ok is false for an empty slice.
func Range(vals []float64) (lo, hi float64, ok bool) {
	if len(vals) == 0 {
		return 0, 0, false
	}
	return slices.Min(vals), slices.Max(vals), true
}

// RateTracker derives a per-second rate from a monotonic counter such as
// Postgres xact_commit + xact_rollback. A counter that goes backwards (server
// restart or stats reset) restarts the baseline instead of yielding a negative
// rate.
type RateTracker struct {
	last   uint64
	at     time.Time
	primed bool
}

// Update records counter at time at and returns the rate since the previous
// call. ok is false for the first sample, after a reset, or when no time passed.
func (t *RateTracker) Update(at time.Time, counter uint64) (rate float64, ok bool) {
	defer func() { t.last, t.at, t.primed = counter, at, true }()
	if !t.primed || counter < t.last {
		return 0, false
	}
	dt := at.Sub(t.at).Seconds()
	if dt <= 0 {
		return 0, false
	}
	return float64(counter-t.last) / dt, true
}

// Reset forgets the baseline.
func (t *RateTracker) Reset() { *t = RateTracker{} }

// LatencyTracker keeps the most recent query durations in a fixed window and
// answers percentile questions. Safe for concurrent use.
type LatencyTracker struct {
	mu    sync.Mutex
	ring  *MetricRing[time.Duration]
	total uint64
}

// DefaultLatencyWindow is the number of recent queries percentiles cover.
const DefaultLatencyWindow = 512

// NewLatencyTracker returns a tracker over the last window durations.
func NewLatencyTracker(window int) *LatencyTracker {
	if window <= 0 {
		window = DefaultLatencyWindow
	}
	return &LatencyTracker{ring: NewRing[time.Duration](window)}
}

// Observe records one query duration.
func (l *LatencyTracker) Observe(d time.Duration) {
	l.mu.Lock()
	l.ring.Push(d)
	l.total++
	l.mu.Unlock()
}

// Count returns the number of queries observed since creation.
func (l *LatencyTracker) Count() uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.total
}

// Percentile returns the nearest-rank q-th percentile (0 < q <= 100) of the
// window, or 0 when empty.
func (l *LatencyTracker) Percentile(q float64) time.Duration {
	l.mu.Lock()
	vals := l.ring.Snapshot()
	l.mu.Unlock()
	return percentile(vals, q)
}

// Percentile constants: the median and tail quantiles, the percent scale, and
// the epsilon that turns a truncating int conversion into a ceiling.
const (
	percentileMedian = 50
	percentileTail   = 95
	percentScale     = 100
	ceilEpsilon      = 0.999999999
)

// P50 and P95 are the median and 95th percentile.
func (l *LatencyTracker) P50() time.Duration { return l.Percentile(percentileMedian) }

// P95 returns the 95th percentile.
func (l *LatencyTracker) P95() time.Duration { return l.Percentile(percentileTail) }

func percentile(vals []time.Duration, q float64) time.Duration {
	if len(vals) == 0 {
		return 0
	}
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
	q = max(0, min(q, percentScale))
	rank := int(q/percentScale*float64(len(vals)) + ceilEpsilon)
	rank = max(1, min(rank, len(vals)))
	return vals[rank-1]
}

// SlowQuery is one captured slow statement.
type SlowQuery struct {
	At       time.Time
	Duration time.Duration
	Rows     int64
	SQL      string
	Source   string // "quarry" for queries run in the editor, "server" for pg_stat_activity
}

// DefaultSlowThreshold is the default capture threshold.
const DefaultSlowThreshold = 50 * time.Millisecond

// SlowLog captures queries at or above a threshold in a bounded ring. Safe for
// concurrent use.
type SlowLog struct {
	mu        sync.Mutex
	threshold time.Duration
	ring      *MetricRing[SlowQuery]
}

// NewSlowLog returns a log keeping the last capacity entries (DefaultCapacity
// when <= 0) at a threshold (DefaultSlowThreshold when <= 0).
func NewSlowLog(threshold time.Duration, capacity int) *SlowLog {
	if threshold <= 0 {
		threshold = DefaultSlowThreshold
	}
	return &SlowLog{threshold: threshold, ring: NewRing[SlowQuery](capacity)}
}

// Threshold returns the current capture threshold.
func (s *SlowLog) Threshold() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.threshold
}

// SetThreshold changes the threshold for future observations.
func (s *SlowLog) SetThreshold(d time.Duration) {
	s.mu.Lock()
	s.threshold = max(d, 0)
	s.mu.Unlock()
}

// Observe captures q when it ran at least as long as the threshold and reports
// whether it was captured.
func (s *SlowLog) Observe(q SlowQuery) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if q.Duration < s.threshold {
		return false
	}
	s.ring.Push(q)
	return true
}

// Entries returns captured queries, newest first.
func (s *SlowLog) Entries() []SlowQuery {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.ring.Snapshot()
	slices.Reverse(out)
	return out
}

// Clear empties the log.
func (s *SlowLog) Clear() {
	s.mu.Lock()
	s.ring.Reset()
	s.mu.Unlock()
}
