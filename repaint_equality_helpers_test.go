package ownframe

import (
	"context"
	"testing"
)

// newPage builds a page, lets the case set its data and handlers, and lays
// it out once.
func newPage(t *testing.T, cfg Config, setup func(p *Page)) *Page {
	t.Helper()

	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if setup != nil {
		setup(p)
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func clickID(t *testing.T, p *Page, ctx context.Context, id string) {
	t.Helper()

	x, y := boxCenter(t, p, id)
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}

func boxCenter(t *testing.T, p *Page, id string) (float64, float64) {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID == id {
			return b.X + b.W/2, b.Y + b.H/2
		}
	}

	t.Fatalf("no box %q", id)

	return 0, 0
}
