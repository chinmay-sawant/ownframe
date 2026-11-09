package page_test

import (
	"context"
	"testing"
)

func TestDisplayReleaseOnCommit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<div id="box">first</div>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after the first redraw")
	}

	// Host takes the dirty rects, setting p.painted and allowing maybeReleasePainted to clear it.
	p.TakeDirtyRects()

	// Update content to force a fresh layout.
	if err := p.Load(ctx, `<div id="box">second</div>`); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	second := p.Display()
	if second == nil {
		t.Fatal("no display after second redraw")
	}
	if second == first {
		t.Fatal("expected second display to differ from first")
	}

	// first should have been released during commit; calling Release again is safe and idempotent.
	first.Release()
}
