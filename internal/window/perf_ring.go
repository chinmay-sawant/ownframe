package window

import (
	"slices"
	"time"
)

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
