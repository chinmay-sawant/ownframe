package page_test

import (
	"context"
	"testing"
)

// TestRedrawBatchDrawsFocusAndClickOnce checks the pointer batch: a focus
// change and the click that follows draw one frame, and the frame keeps the
// focus.
func TestRedrawBatchDrawsFocusAndClickOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<style>#go:focus{color:#ff0000}</style><button id="go">Start</button>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	x, y := boxCenter(t, p, "go")
	before := p.Stats().Layouts

	p.BeginRedrawBatch()
	if err := p.Focus(ctx, "go"); err != nil {
		t.Fatal(err)
	}
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Layouts; got != before+1 {
		t.Fatalf("layouts = %d, want one batch draw", got-before)
	}
	if got := p.FocusID(); got != "go" {
		t.Fatalf("focus = %q, want go", got)
	}
}

// TestRedrawBatchSkipsUnchangedFlush checks that the flush runs the redraw
// sequence once but lays nothing out when the source and state are unchanged,
// whether the batch held a plain Redraw or only RedrawIfChanged calls.
func TestRedrawBatchSkipsUnchangedFlush(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>same</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats()

	p.BeginRedrawBatch()
	if err := p.RedrawIfChanged(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats(); got.Layouts != before.Layouts || got.Repaints != before.Repaints {
		t.Fatalf("layouts = %d, repaints = %d, want no draw", got.Layouts-before.Layouts, got.Repaints-before.Repaints)
	}
	if got := p.Stats().Redraws; got != before.Redraws+1 {
		t.Fatalf("redraws = %d, want one flush", got-before.Redraws)
	}
}
