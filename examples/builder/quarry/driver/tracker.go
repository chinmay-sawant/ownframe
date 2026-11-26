package driver

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// tracker records in-flight statements issued through a Driver so SQLite (and
// the mock) can report ActiveQueries and honor CancelQuery.
type tracker struct {
	mu       sync.Mutex
	next     int
	inflight map[int]*flight
	total    int64
	last     time.Duration
	latency  []time.Duration // ring of recent durations
	user     string
	database string
}

type flight struct {
	q      ActiveQuery
	cancel context.CancelFunc
}

const latencyRing = 128

func newTracker(user, database string) *tracker {
	return &tracker{inflight: map[int]*flight{}, user: user, database: database}
}

// begin registers a statement and returns a derived context plus a finish
// function that records latency.
func (t *tracker) begin(ctx context.Context, sqlText string) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	t.mu.Lock()
	t.next++
	pid := t.next
	now := time.Now()
	t.inflight[pid] = &flight{
		q: ActiveQuery{
			PID: pid, User: t.user, Database: t.database, State: "active",
			Query: sqlText, StartedAt: now,
		},
		cancel: cancel,
	}
	t.mu.Unlock()
	return ctx, func() {
		d := time.Since(now)
		t.mu.Lock()
		delete(t.inflight, pid)
		t.total++
		t.last = d
		if len(t.latency) < latencyRing {
			t.latency = append(t.latency, d)
		} else {
			t.latency[int(t.total)%latencyRing] = d
		}
		t.mu.Unlock()
		cancel()
	}
}

func (t *tracker) cancel(pid int) error {
	t.mu.Lock()
	f, ok := t.inflight[pid]
	t.mu.Unlock()
	if !ok {
		return fmt.Errorf("driver: no running statement with pid %d", pid)
	}
	f.cancel()
	return nil
}

func (t *tracker) active() []ActiveQuery {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	out := make([]ActiveQuery, 0, len(t.inflight))
	for _, f := range t.inflight {
		q := f.q
		q.Duration = now.Sub(q.StartedAt)
		out = append(out, q)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PID < out[j].PID })
	return out
}

func (t *tracker) totals() (statements int64, lastMs, avgMs float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	statements = t.total
	if len(t.latency) == 0 {
		return
	}
	var sum time.Duration
	for _, d := range t.latency {
		sum += d
	}
	avgMs = float64(sum) / float64(len(t.latency)) / float64(time.Millisecond)
	lastMs = float64(t.last) / float64(time.Millisecond)
	return
}
