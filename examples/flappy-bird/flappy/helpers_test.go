package flappy

import (
	"context"
	"math/rand/v2"
	"testing"
	"time"
)

// testClock is a clock the test moves by hand.
type testClock struct {
	at time.Time
}

func (c *testClock) now() time.Time {
	return c.at
}

func (c *testClock) advance(d time.Duration) {
	c.at = c.at.Add(d)
}

// newTestApp returns a page with a seeded generator and a stopped clock.
func newTestApp(t *testing.T) (*App, *testClock) {
	t.Helper()

	app, err := New()
	if err != nil {
		t.Fatal(err)
	}

	clock := &testClock{at: time.Unix(0, 0)}
	app.now = clock.now
	app.rng = rand.New(rand.NewPCG(1, 2)) //nolint:gosec // reason: deterministic test clock seed, not security.

	if err := app.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return app, clock
}

// tick advances the clock by frame and runs one game frame.
func tick(t *testing.T, app *App, clock *testClock, frame time.Duration) {
	t.Helper()

	clock.advance(frame)

	if err := app.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// press sends a key press through the page handlers.
func press(t *testing.T, app *App, key string) {
	t.Helper()

	if err := app.page.Testing().KeyDown(context.Background(), key); err != nil {
		t.Fatal(err)
	}
}
