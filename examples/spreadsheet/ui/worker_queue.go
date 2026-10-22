package ui

import (
	"sync"
	"time"
)

// worker runs backend calls off the UI loop. Fetches are disposable and
// coalesced; commands and file reads are never dropped.
type worker struct {
	b      Backend
	mu     sync.Mutex
	queue  []job
	busy   bool
	fetch  uint64
	closed bool

	wake chan struct{}
	stop chan struct{}
	done chan struct{}
	out  chan result
}

func newWorker(b Backend) *worker {
	w := &worker{
		b:    b,
		wake: make(chan struct{}, 1),
		stop: make(chan struct{}),
		done: make(chan struct{}),
		out:  make(chan result, uiQueueCap),
	}
	go w.run()

	return w
}

// post queues a job. A new fetch replaces queued fetches, so a stale range
// load never runs. It never blocks.
func (w *worker) post(j job) bool {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()

		return false
	}

	if j.kind == jobFetch {
		w.fetch = j.gen
		kept := w.queue[:0]
		for _, q := range w.queue {
			if q.kind != jobFetch {
				kept = append(kept, q)
			}
		}

		kept = append(kept, j)
		w.queue = kept
	} else {
		w.queue = append(w.queue, j)
	}
	w.mu.Unlock()

	select {
	case w.wake <- struct{}{}:
	default:
	}

	return true
}

// idle reports that no job is queued or running.
func (w *worker) idle() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	return len(w.queue) == 0 && !w.busy
}

// close stops the worker and reports whether it joined within budget.
func (w *worker) close(budget time.Duration) bool {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()

		return true
	}

	w.closed = true
	w.mu.Unlock()
	close(w.stop)

	select {
	case <-w.done:
		return true
	case <-time.After(budget):
		return false
	}
}
