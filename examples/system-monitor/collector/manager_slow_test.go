package collector

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestManagerSlowSourceSkips checks that a source slower than its deadline
// cannot pile up work: ticks are skipped and Close still finishes inside a
// budget.
func TestManagerSlowSourceSkips(t *testing.T) {
	src := &fakeSource{slow: 60 * time.Millisecond}
	m := New(Options{
		Source:          src,
		SummaryInterval: 10 * time.Millisecond,
		ProcessInterval: 10 * time.Millisecond,
		DetailInterval:  10 * time.Millisecond,
		Deadline:        25 * time.Millisecond,
	})
	if err := m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}

	time.Sleep(250 * time.Millisecond)

	if stats := m.Stats(); stats.Skipped == 0 {
		t.Fatalf("no skipped ticks: %+v", stats)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := m.Close(ctx); err != nil {
		t.Fatalf("close = %v", err)
	}
}

// TestManagerCloseTwice checks that Close is idempotent and that a manager
// without Start closes immediately.
func TestManagerCloseTwice(t *testing.T) {
	m := New(Options{Source: &fakeSource{}})

	if err := m.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := m.Start(t.Context()); !errors.Is(err, ErrClosed) {
		t.Fatalf("start after close = %v", err)
	}
}
