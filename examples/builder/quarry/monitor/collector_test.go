package monitor

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func settledGoroutines(t *testing.T, want int) {
	t.Helper()
	waitFor(t, "goroutines to exit", func() bool { return runtime.NumGoroutine() <= want })
}

func TestCollectorPollsAndDrains(t *testing.T) {
	var n atomic.Int64
	c := NewCollector(func(context.Context) (int, error) { return int(n.Add(1)), nil }, 5*time.Millisecond)
	c.Start(context.Background())
	defer c.Stop()
	var got []Result[int]
	waitFor(t, "3 samples", func() bool {
		got = append(got, c.Drain()...)
		return len(got) >= 3
	})
	for i := 1; i < len(got); i++ {
		if got[i].Sample <= got[i-1].Sample || got[i].At.IsZero() {
			t.Fatalf("order: %+v", got)
		}
	}
	if !c.Running() || c.Interval() != 5*time.Millisecond {
		t.Fatal("state")
	}
}

func TestCollectorErrorsAreResults(t *testing.T) {
	boom := errors.New("boom")
	c := NewCollector(func(context.Context) (int, error) { return 0, boom }, 0)
	c.Start(context.Background())
	defer c.Stop()
	c.RefreshNow()
	var r []Result[int]
	waitFor(t, "error result", func() bool { r = c.Drain(); return len(r) > 0 })
	if !errors.Is(r[0].Err, boom) {
		t.Fatal(r[0])
	}
}

func TestCollectorPauseAndRefreshNow(t *testing.T) {
	var n atomic.Int64
	c := NewCollector(func(context.Context) (int, error) { return int(n.Add(1)), nil }, 0)
	c.Start(context.Background())
	defer c.Stop()
	time.Sleep(40 * time.Millisecond)
	if n.Load() != 0 || c.Pending() {
		t.Fatalf("paused collector polled %d times", n.Load())
	}
	c.RefreshNow()
	waitFor(t, "manual refresh", func() bool { return n.Load() == 1 })
	time.Sleep(30 * time.Millisecond)
	if n.Load() != 1 {
		t.Fatal("paused collector kept polling")
	}
	// Resume then pause again through SetInterval.
	c.SetInterval(3 * time.Millisecond)
	waitFor(t, "resumed polling", func() bool { return n.Load() >= 4 })
	c.SetInterval(0)
	time.Sleep(20 * time.Millisecond) // let an in-flight tick land
	settled := n.Load()
	time.Sleep(40 * time.Millisecond)
	if n.Load() != settled {
		t.Fatalf("still polling after pause: %d -> %d", settled, n.Load())
	}
}

func TestCollectorBufferBoundedKeepsNewest(t *testing.T) {
	var n atomic.Int64
	c := NewCollector(func(context.Context) (int, error) { return int(n.Add(1)), nil }, time.Millisecond)
	c.Start(context.Background())
	defer c.Stop()
	waitFor(t, "buffer overrun", func() bool { return n.Load() > 3*ResultBuffer })
	c.SetInterval(0)
	time.Sleep(20 * time.Millisecond)
	got := c.Drain()
	if len(got) != ResultBuffer {
		t.Fatalf("buffered %d, want exactly %d", len(got), ResultBuffer)
	}
	if got[len(got)-1].Sample != int(n.Load()) {
		t.Fatalf("newest sample dropped: last %d of %d", got[len(got)-1].Sample, n.Load())
	}
	if len(c.Drain()) != 0 || c.Pending() {
		t.Fatal("drain must empty the buffer")
	}
}

func TestCollectorStopCancelsInFlightAndLeaksNothing(t *testing.T) {
	runtime.GC()
	base := runtime.NumGoroutine()
	started := make(chan struct{}, 1)
	c := NewCollector(func(ctx context.Context) (int, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		<-ctx.Done() // a hung server
		return 0, ctx.Err()
	}, time.Millisecond)
	c.Start(context.Background())
	c.Start(context.Background()) // second Start is a no-op
	<-started
	done := make(chan struct{})
	go func() { c.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop blocked on an in-flight fetch")
	}
	if c.Running() {
		t.Fatal("still running")
	}
	if len(c.Drain()) != 0 {
		t.Fatal("cancelled fetch must not be reported as a sample")
	}
	c.Stop() // idempotent
	settledGoroutines(t, base)
}

func TestCollectorParentContextCancel(t *testing.T) {
	runtime.GC()
	base := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	c := NewCollector(func(context.Context) (int, error) { return 1, nil }, 2*time.Millisecond)
	c.Start(ctx)
	cancel()
	settledGoroutines(t, base)
	c.Stop()
}

func TestCollectorSwitchConnection(t *testing.T) {
	// A fresh collector per connection: stop the old, start the new, no overlap.
	runtime.GC()
	base := runtime.NumGoroutine()
	for i := 0; i < 20; i++ {
		c := NewCollector(func(context.Context) (int, error) { return i, nil }, time.Millisecond)
		c.Start(context.Background())
		c.RefreshNow()
		c.SetInterval(2 * time.Millisecond)
		c.Stop()
	}
	settledGoroutines(t, base)
}
