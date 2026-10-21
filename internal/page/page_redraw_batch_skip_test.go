package page_test

import (
	"context"
	"testing"
)

// TestRedrawBatchSkipsUnchangedIfChanged checks that an IfChanged-only batch
// still honors the unchanged check at flush time.
func TestRedrawBatchSkipsUnchangedIfChanged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>same</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Layouts

	p.BeginRedrawBatch()
	if err := p.RedrawIfChanged(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Layouts; got != before {
		t.Fatalf("layouts = %d, want no draw", got-before)
	}
}
