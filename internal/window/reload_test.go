package window

import (
	"context"
	"testing"
	"time"
)

func TestHotReloadPollsOnTheInterval(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.left = 1

	s := newShell(context.Background(), app)
	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 1 {
		t.Fatalf("polls = %d, want 1", app.polls)
	}

	if app.redraws != 2 {
		t.Fatalf("redraws = %d, want one reload", app.redraws)
	}

	if s.display == nil || s.seq != 2 {
		t.Fatalf("display = %v, seq = %d, want the new generation", s.display, s.seq)
	}
}

func TestHotReloadWaitsForTheInterval(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.left = 1

	s := newShell(context.Background(), app)
	s.lastPoll = time.Now()

	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 0 {
		t.Fatalf("polls = %d, want 0 inside the interval", app.polls)
	}

	s.lastPoll = time.Now().Add(-time.Second)

	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 1 {
		t.Fatalf("polls = %d, want 1 after the interval", app.polls)
	}
}
