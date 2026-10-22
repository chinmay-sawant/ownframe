package collector

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestPoolQueueBounded checks that a full queue drops instead of blocking and
// that the drop is counted.
func TestPoolQueueBounded(t *testing.T) {
	p := NewPool(1, 1, time.Second)
	defer p.Close()

	release := make(chan struct{})
	started := make(chan struct{}, 1)
	block := func(ctx context.Context) error {
		started <- struct{}{}
		<-release

		return nil
	}

	if !p.TryDo(context.Background(), block, nil) {
		t.Fatal("first call not queued")
	}
	<-started

	if !p.TryDo(context.Background(), block, nil) {
		t.Fatal("second call not queued")
	}
	if p.TryDo(context.Background(), block, nil) {
		t.Fatal("third call queued past the bound")
	}

	close(release)
	p.Close()

	if stats := p.Stats(); stats.Dropped != 1 || stats.Completed < 2 {
		t.Fatalf("stats = %+v", stats)
	}
}

// TestPoolDeadline checks that a call which ignores its context is counted as
// timed out once the deadline fires.
func TestPoolDeadline(t *testing.T) {
	p := NewPool(1, 1, 30*time.Millisecond)
	defer p.Close()

	done := make(chan error, 1)
	if !p.TryDo(context.Background(), func(ctx context.Context) error {
		<-ctx.Done()

		return ctx.Err()
	}, func(err error) { done <- err }) {
		t.Fatal("call not queued")
	}

	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("deadline never fired")
	}

	if stats := p.Stats(); stats.TimedOut != 1 {
		t.Fatalf("stats = %+v", stats)
	}
}

// TestPoolDoAfterClose checks that Do reports a full pool rather than
// panicking after Close.
func TestPoolDoAfterClose(t *testing.T) {
	p := NewPool(1, 1, time.Second)
	p.Close()

	if err := p.Do(context.Background(), func(context.Context) error { return nil }); !errors.Is(err, ErrPoolFull) {
		t.Fatalf("err = %v", err)
	}
}
