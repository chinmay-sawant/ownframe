package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestZZProbeFixed(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	count := func(tag string) {
		d := a.page.Probe().Display()
		fixed, full := 0, 0
		for _, op := range d.Ops {
			if op.Fixed {
				fixed++
				if op.W >= 1400 && op.H >= 900 {
					full++
					t.Logf("%s fullscreen fixed op: x=%v y=%v w=%v h=%v", tag, op.X, op.Y, op.W, op.H)
				}
			}
		}
		t.Logf("%s: fixed=%d fullscreen-fixed=%d", tag, fixed, full)
	}
	count("closed")
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	count("search-open")
}
