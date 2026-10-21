package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestClickPaintPlainClick checks that a click on a plain box with a paint
// handler skips the redraw too.
func TestClickPaintPlainClick(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := newCachePage(t, `<div id="d">{{.N}}</div>`)
	screen.SetData(struct{ N int }{1})
	screen.Handle(page.Handlers{
		ClickPaint: func(context.Context, page.Box) (bool, error) { return true, nil },
	})
	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Stats()
	gen := screen.Generation()
	clickID(t, screen, "d")

	if got := screen.Stats().Redraws; got != before.Redraws {
		t.Fatalf("redraws = %d, want %d", got, before.Redraws)
	}
	if got := screen.Generation(); got != gen {
		t.Fatalf("generation = %d, want %d", got, gen)
	}
}

// TestClickPaintFalseRedraws checks that a click whose paint handler does
// not paint still redraws.
func TestClickPaintFalseRedraws(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<button id="go">Go</button>`)
	screen.Handle(page.Handlers{
		ClickPaint: func(context.Context, page.Box) (bool, error) { return false, nil },
	})

	before := screen.Stats()
	gen := screen.Generation()
	clickID(t, screen, "go")

	if got := screen.Stats().Redraws; got != before.Redraws+1 {
		t.Fatalf("redraws = %d, want %d", got, before.Redraws+1)
	}
	if got := screen.Generation(); got <= gen {
		t.Fatalf("generation = %d, want past %d", got, gen)
	}
}
