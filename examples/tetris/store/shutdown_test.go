package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCloseDrainsQueuedRequests(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	done := make(chan error, 8)

	for i := 0; i < 8; i++ {
		go func() { done <- s.SaveSettings(ctx, DefaultSettings()) }()
	}

	time.Sleep(10 * time.Millisecond)

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 8; i++ {
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, ErrClosed) {
				t.Fatalf("queued write returned %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("a queued write never returned")
		}
	}
}

func TestCanceledContextFailsFast(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := s.SaveSettings(ctx, DefaultSettings()); err == nil {
		t.Fatal("a canceled context was accepted")
	}
}

func TestCloseWhileWritesAreInFlight(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 4)

	for i := 0; i < 4; i++ {
		go func(i int) {
			done <- s.SaveGame(context.Background(), testResult(string(rune('a'+i)), 100), nil) //nolint:gosec // reason: test builds a short id from a small index, the rune stays in range.
		}(i)
	}

	time.Sleep(5 * time.Millisecond)

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, ErrClosed) {
				t.Fatalf("in-flight write returned %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("an in-flight write never returned")
		}
	}
}
