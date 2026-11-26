package monitor

import (
	"context"
	"sync"
	"time"
)

// Result is one poll outcome delivered to the UI.
type Result[S any] struct {
	At     time.Time
	Sample S
	Err    error
}

// FetchFunc collects one sample, honoring ctx cancellation. For a driver this
// wraps Driver.CollectMetrics.
type FetchFunc[S any] func(ctx context.Context) (S, error)

// ResultBuffer is the capacity of the bounded result channel.
const ResultBuffer = 8

// Collector polls fetch on a background goroutine and hands results to the UI
// through a bounded buffer that Drain empties from SetTick. When the UI falls
// behind, the oldest undrained result is dropped, so memory stays bounded and
// the newest data always survives. An interval of 0 pauses polling; RefreshNow
// still works while paused.
type Collector[S any] struct {
	fetch FetchFunc[S]
	now   func() time.Time

	out     chan Result[S]
	refresh chan struct{}
	ctl     chan time.Duration

	mu       sync.Mutex
	interval time.Duration
	cancel   context.CancelFunc
	done     chan struct{}
}

// NewCollector creates a stopped collector with the given initial interval.
func NewCollector[S any](fetch FetchFunc[S], interval time.Duration) *Collector[S] {
	return &Collector[S]{
		fetch:    fetch,
		now:      time.Now,
		out:      make(chan Result[S], ResultBuffer),
		refresh:  make(chan struct{}, 1),
		ctl:      make(chan time.Duration, 1),
		interval: max(interval, 0),
	}
}

// Start launches the polling goroutine, which exits when ctx is cancelled or
// Stop is called. Starting a running collector is a no-op. It polls once
// immediately unless paused.
func (c *Collector[S]) Start(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.done = make(chan struct{})
	go func(done chan struct{}, every time.Duration) {
		defer cancel()
		c.loop(ctx, done, every)
	}(c.done, c.interval)
}

// Stop cancels polling (including an in-flight fetch) and waits for the
// goroutine to exit. Safe to call repeatedly or without Start.
func (c *Collector[S]) Stop() {
	c.mu.Lock()
	cancel, done := c.cancel, c.done
	c.cancel, c.done = nil, nil
	c.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	<-done
}

// Running reports whether the goroutine is active.
func (c *Collector[S]) Running() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cancel != nil
}

// Interval returns the current poll interval (0 when paused).
func (c *Collector[S]) Interval() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.interval
}

// SetInterval changes the poll interval; 0 pauses. It takes effect promptly.
func (c *Collector[S]) SetInterval(d time.Duration) {
	d = max(d, 0)
	c.mu.Lock()
	c.interval = d
	running := c.cancel != nil
	c.mu.Unlock()
	if !running {
		return
	}
	select { // replace any unread interval change with the newest
	case <-c.ctl:
	default:
	}
	select {
	case c.ctl <- d:
	default:
	}
}

// RefreshNow requests an immediate poll. Requests coalesce.
func (c *Collector[S]) RefreshNow() {
	select {
	case c.refresh <- struct{}{}:
	default:
	}
}

// Drain returns every buffered result, oldest first, without blocking.
func (c *Collector[S]) Drain() []Result[S] {
	var out []Result[S]
	for {
		select {
		case r := <-c.out:
			out = append(out, r)
		default:
			return out
		}
	}
}

// Pending reports whether Drain would return anything, for idle-frame decisions.
func (c *Collector[S]) Pending() bool { return len(c.out) > 0 }

func (c *Collector[S]) loop(ctx context.Context, done chan struct{}, interval time.Duration) {
	defer close(done)
	var timer *time.Timer
	var tick <-chan time.Time
	arm := func(d time.Duration) {
		if timer != nil {
			timer.Stop()
		}
		timer, tick = nil, nil
		if d > 0 {
			timer = time.NewTimer(d)
			tick = timer.C
		}
	}
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	if interval > 0 {
		c.poll(ctx)
	}
	arm(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case d := <-c.ctl:
			interval = d
			if d > 0 {
				c.poll(ctx)
			}
			arm(d)
		case <-c.refresh:
			c.poll(ctx)
			arm(interval)
		case <-tick:
			c.poll(ctx)
			arm(interval)
		}
	}
}

func (c *Collector[S]) poll(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	s, err := c.fetch(ctx)
	if ctx.Err() != nil {
		return // cancelled mid-flight: do not report the cancellation as a sample
	}
	r := Result[S]{At: c.now(), Sample: s, Err: err}
	for {
		select {
		case c.out <- r:
			return
		default:
		}
		select { // buffer full: drop the oldest
		case <-c.out:
		default:
		}
	}
}
