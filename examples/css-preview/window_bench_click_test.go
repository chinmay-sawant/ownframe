package main

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// windowBenchClick sends the desktop sequence used by the real walk:
// press, focus, and click in one redraw batch, then release outside it.
func windowBenchClick(ctx context.Context, p *ownframe.Page, tb testing.TB, id string) time.Duration {
	tb.Helper()
	box, ok := findBox(p.Boxes(), id)
	if !ok {
		tb.Fatalf("missing %s", id)
	}
	x, y := box.X+box.W/2, box.Y+box.H/2
	if err := p.Hover(ctx, x, y); err != nil {
		tb.Fatal(err)
	}
	before := p.Stats()
	err := redrawBatch(ctx, p, func() error {
		if err := p.Press(ctx, x, y); err != nil {
			return err
		}
		if err := p.Testing().SelectAt(ctx, x, y); err != nil {
			return err
		}
		return p.Click(ctx, x, y)
	})
	after := p.Stats()
	if err == nil {
		err = p.Release(ctx)
	}
	if err != nil {
		tb.Fatal(err)
	}
	if after.Redraws != before.Redraws+1 || after.Layouts != before.Layouts+1 {
		tb.Fatalf("%s: redraws %d->%d layouts %d->%d, want one draw",
			id, before.Redraws, after.Redraws, before.Layouts, after.Layouts)
	}
	return after.LastRedraw
}
