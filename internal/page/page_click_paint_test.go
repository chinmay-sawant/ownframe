package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestClickPaintSkipsRedraw checks that a click whose paint handler says it
// painted leaves the generation and the redraw counter alone, keeps the
// focus, and lets a later state change redraw again.
func TestClickPaintSkipsRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := newCachePage(t, `<button id="go">{{.N}}</button>`)
	screen.SetData(struct{ N int }{1})
	calls := 0
	screen.Handle(page.Handlers{
		ClickPaint: func(context.Context, page.Box) (bool, error) {
			calls++
			return calls == 1, nil
		},
	})
	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Stats()
	gen := screen.Generation()
	clickID(t, screen, "go")

	after := screen.Stats()
	if after.Redraws != before.Redraws {
		t.Fatalf("redraws = %d, want %d", after.Redraws, before.Redraws)
	}
	if after.Layouts != before.Layouts {
		t.Fatalf("layouts = %d, want %d", after.Layouts, before.Layouts)
	}
	if got := screen.Generation(); got != gen {
		t.Fatalf("generation = %d, want %d", got, gen)
	}
	if got := screen.FocusID(); got != "go" {
		t.Fatalf("focus = %q, want go", got)
	}

	// A later event that changes the state redraws again.
	screen.SetData(struct{ N int }{2})
	clickID(t, screen, "go")

	if got := screen.Stats().Redraws; got != before.Redraws+1 {
		t.Fatalf("redraws = %d, want %d", got, before.Redraws+1)
	}
	if got := screen.Generation(); got <= gen {
		t.Fatalf("generation = %d, want past %d", got, gen)
	}
}
