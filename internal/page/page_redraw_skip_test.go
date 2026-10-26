package page_test

import (
	"context"
	"testing"
)

// TestRedrawSkipsIdenticalFrame checks that a second Redraw with no input
// change counts the redraw but lays nothing out again: no new layout,
// repaint, parse, or dirty region, the same retained display list, and no
// layout time.
func TestRedrawSkipsIdenticalFrame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>stable</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after the first redraw")
	}
	p.TakeDirtyRects()
	before := p.Stats()
	gen := p.Generation()

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	after := p.Stats()
	if after.Redraws != before.Redraws+1 {
		t.Fatalf("redraws = %d, want %d", after.Redraws, before.Redraws+1)
	}
	if after.Layouts != before.Layouts || after.Repaints != before.Repaints {
		t.Fatalf("layouts = %d, repaints = %d, want no new work",
			after.Layouts-before.Layouts, after.Repaints-before.Repaints)
	}
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("parses = %d, cascades = %d, want no reparse",
			after.Parses-before.Parses, after.Cascades-before.Cascades)
	}
	if after.StyleTime != 0 || after.LayoutTime != 0 {
		t.Fatalf("style = %v, layout = %v, want both 0", after.StyleTime, after.LayoutTime)
	}
	if p.Generation() != gen {
		t.Fatalf("generation = %d, want %d", p.Generation(), gen)
	}
	if p.Display() != first {
		t.Fatal("identical redraw rebuilt the display list")
	}
	if _, ok := p.TakeDirtyRects(); ok {
		t.Fatal("identical redraw reported dirty")
	}
}

// TestRedrawDrawsChangedData checks the other side of the skip: data that
// changes the executed source still commits a new frame.
func TestRedrawDrawsChangedData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>{{.V}}</p>`)
	p.SetData(struct{ V string }{V: "a"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats()
	gen := p.Generation()

	p.SetData(struct{ V string }{V: "b"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats(); got.Layouts != before.Layouts+1 || got.Repaints != before.Repaints+1 {
		t.Fatalf("layouts = %d, repaints = %d, want one new frame",
			got.Layouts-before.Layouts, got.Repaints-before.Repaints)
	}
	if p.Generation() == gen {
		t.Fatal("changed data did not bump the generation")
	}
}
