package pet

import (
	"context"
	"math/rand/v2"
	"testing"
	"time"
)

type tclock struct {
	at time.Time
}

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
	c := &tclock{at: time.Unix(1000, 0)}
	a.now = c.now
	a.rng = rand.New(rand.NewPCG(42, 42))
	a.state = newState(a.now)
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
