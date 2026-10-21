package page

import (
	"context"
	"testing"
)

// afterRedrawHTML is one filled box, so the page keeps a display list and
// the commit path runs.
const afterRedrawHTML = `<style>div { width: 40px; height: 20px; background: #ff0000; }</style><div id="a"></div>`

// afterRedrawPage builds a page for the after-redraw tests.
func afterRedrawPage(t *testing.T, html string) *Page {
	t.Helper()

	p, err := New(Config{HTML: html, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

// boxWithID returns the box the id names.
func boxWithID(boxes []Box, id string) (Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}

	return Box{}, false
}

func TestAfterRedrawRunsAfterTheCommit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := afterRedrawPage(t, afterRedrawHTML)

	calls := 0
	var display *Display
	var boxes []Box
	p.SetAfterRedraw(func() {
		calls++
		display = p.Display()
		boxes = p.Boxes()
	})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if display == nil {
		t.Fatal("the callback did not see the new display")
	}
	if _, ok := boxWithID(boxes, "a"); !ok {
		t.Fatalf("boxes = %+v, want the box of a", boxes)
	}
}

func TestAfterRedrawSkipsAnUnchangedRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := afterRedrawPage(t, afterRedrawHTML)

	calls := 0
	p.SetAfterRedraw(func() { calls++ })

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d after the first redraw, want 1", calls)
	}

	// The same frame again: nothing changed, so the page skips the layout
	// and the callback stays away.
	if err := p.RedrawIfChanged(ctx); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d after an unchanged redraw, want 1", calls)
	}
}
