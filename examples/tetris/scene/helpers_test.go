package scene

import (
	"context"
	"testing"
	"time"
)

// testClock is a clock the test moves by hand.
type testClock struct{ at time.Time }

func (c *testClock) now() time.Time { return c.at }

func (c *testClock) add(d time.Duration) { c.at = c.at.Add(d) }

// fakeStepper returns queued fixed steps and counts resets.
type fakeStepper struct {
	advance int
	resets  int
}

func (s *fakeStepper) Advance(time.Time) int {
	n := s.advance
	s.advance = 0

	return n
}

func (s *fakeStepper) Reset(time.Time) {
	s.resets++
	s.advance = 0
}

// newTestScene builds a scene on a stopped clock and draws it once.
func newTestScene(t *testing.T, m Model, st Store, opts Options) (*Scene, *testClock) {
	t.Helper()

	clock := &testClock{at: time.Unix(1000, 0)}
	if opts.Now == nil {
		opts.Now = clock.now
	}

	s, err := New(m, st, opts)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return s, clock
}

func tickAt(t *testing.T, s *Scene, clock *testClock, d time.Duration) {
	t.Helper()

	clock.add(d)

	if err := s.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// click finds an element's box and clicks its centre.
func click(t *testing.T, s *Scene, id string) {
	t.Helper()

	box, ok := boxByID(s.page.Boxes(), id)
	if !ok {
		t.Fatalf("no box %q", id)
	}

	if err := s.page.Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}
}
