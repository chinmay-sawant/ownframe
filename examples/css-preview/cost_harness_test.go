package main

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// costFullCatalog renders the whole 43 row catalog. It is the shape Play
// All used to leave behind: every row in the page, no row window, no
// spacers. The scroll window is replaced so the rows stay whole.
func costFullCatalog(ctx context.Context, p *ownframe.Page, v *view,
	d *demo, tb testing.TB,
) {
	tb.Helper()
	v.Rows = v.Catalog
	v.TopPad, v.BotPad = 0, 0
	d.allPlaying = true
	p.SetScrollWindow(func(int, int) bool { return false })
	p.SetData(v)
	if err := p.Redraw(ctx); err != nil {
		tb.Fatal(err)
	}
}

// costClick sends the click the window sends: a hover, then the press, the
// caret, and the click in one redraw batch, then the release. It mirrors
// windowBenchClick and also reports the bytes the click's own redraw
// allocated, which the page counts when perf is on.
func costClick(ctx context.Context, p *ownframe.Page, tb testing.TB,
	id string,
) (time.Duration, uint64) {
	tb.Helper()
	box, ok := findBox(p.Boxes(), id)
	if !ok {
		tb.Fatalf("missing %s", id)
	}
	x, y := box.X+box.W/2, box.Y+box.H/2
	if err := p.Hover(ctx, x, y); err != nil {
		tb.Fatal(err)
	}
	before := p.Probe().Stats()
	err := redrawBatch(ctx, p, func() error {
		if err := p.Press(ctx, x, y); err != nil {
			return err
		}
		if err := p.Testing().SelectAt(ctx, x, y); err != nil {
			return err
		}

		return p.Click(ctx, x, y)
	})
	if err != nil {
		tb.Fatal(err)
	}
	after := p.Probe().Stats()
	if after.Redraws != before.Redraws+1 || after.Layouts != before.Layouts+1 {
		tb.Fatalf("%s: redraws %d->%d layouts %d->%d, want one draw",
			id, before.Redraws, after.Redraws, before.Layouts, after.Layouts)
	}
	if err := p.Release(ctx); err != nil {
		tb.Fatal(err)
	}

	return after.LastRedraw, after.AllocFrame
}
