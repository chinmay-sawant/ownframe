package page_test

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestStatsCountRedraws checks the Phase 1 exit.
func TestStatsCountRedraws(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p id="known">hello</p>`)

	for range 2 {
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}

	stats := p.Stats()
	if stats.Redraws != 2 || stats.Parses == 0 {
		t.Fatalf("redraws = %d, parses = %d, want 2 and a parse", stats.Redraws, stats.Parses)
	}

	// The second Redraw sees an identical frame and skips the layout, so
	// it counts the redraw but no layout or repaint, and no layout time.
	if stats.Layouts != 1 || stats.Repaints != 1 {
		t.Fatalf("layouts = %d, repaints = %d, want 1 each", stats.Layouts, stats.Repaints)
	}
	if stats.LayoutTime != 0 {
		t.Fatalf("layout time = %v, want 0 on the skipped frame", stats.LayoutTime)
	}

	if stats.Boxes == 0 || stats.Ops == 0 {
		t.Fatalf("boxes = %d, ops = %d", stats.Boxes, stats.Ops)
	}

	if stats.LastRedraw <= 0 || stats.LastDraw <= 0 {
		t.Fatalf("last redraw = %v, last draw = %v", stats.LastRedraw, stats.LastDraw)
	}
}

func TestStatsClickReportsBoxID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<div id="known">pick me</div>`)

	got := ""
	p.Handle(page.Handlers{
		Click: func(_ context.Context, box page.Box) error {
			got = box.ID

			return nil
		},
	})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	x, y := boxCenter(t, p, "known")
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	if got != "known" {
		t.Fatalf("click id = %q, want known", got)
	}
}

// TestSetDrawTimeWinsLastDraw checks the window hook: once a shell records
// its draw time, LastDraw is that time, not the page's own paint stage.
func TestSetDrawTimeWinsLastDraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>frame</p>`)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	p.SetDrawTime(7 * time.Millisecond)
	if got := p.Stats().LastDraw; got != 7*time.Millisecond {
		t.Fatalf("last draw = %v, want 7ms", got)
	}
}
