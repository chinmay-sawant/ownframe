package page_test

import (
	"context"
	"testing"
)

// TestRedrawBatchNestedDrawsOnce checks that a nested batch draws once: the
// inner EndRedrawBatch leaves the draw to the outer one.
func TestRedrawBatchNestedDrawsOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>same</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Layouts

	p.BeginRedrawBatch()
	p.BeginRedrawBatch()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != before {
		t.Fatalf("layouts = %d, want no draw from the inner batch", got-before)
	}

	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != before+1 {
		t.Fatalf("layouts = %d, want one draw from the outer batch", got-before)
	}
}

// TestRedrawBatchQueuesAFailedDraw checks that a draw that fails leaves the
// redraw queued, so the next batch commits it and the frame catches up.
func TestRedrawBatchQueuesAFailedDraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>same</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Layouts

	dead, cancel := context.WithCancel(ctx)
	cancel()

	p.BeginRedrawBatch()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(dead); err == nil {
		t.Fatal("the final draw should fail with a dead context")
	}
	if got := p.Stats().Layouts; got != before {
		t.Fatalf("layouts = %d, want no draw", got-before)
	}

	p.BeginRedrawBatch()
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != before+1 {
		t.Fatalf("layouts = %d, want the queued draw", got-before)
	}
}
