package eyes

import (
	"context"
	"math/rand/v2"
	"testing"
	"time"
)

type tclock struct{ at time.Time }

func (c *tclock) now() time.Time { return c.at }
func (c *tclock) adv(d time.Duration) {
	c.at = c.at.Add(d)
}

func newT(t *testing.T) (*App, *tclock) {
	t.Helper()
	a, err := New()
	if err != nil {
		t.Fatal(err)
	}
	c := &tclock{at: time.Unix(0, 0)}
	a.now = c.now
	a.rng = rand.New(rand.NewPCG(1, 2)) //nolint:gosec // reason: seeded generator makes eye tests deterministic, not security-sensitive.
	a.st = newState(a.now)
	if err := a.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return a, c
}

func tk(t *testing.T, a *App, c *tclock, d time.Duration) {
	t.Helper()
	c.adv(d)
	if err := a.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}
