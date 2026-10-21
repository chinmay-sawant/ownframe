package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestClickPaintKeepsDamage checks that the damage the handler marked stays
// for the window after a paint-only click, so the window repaints it.
func TestClickPaintKeepsDamage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := newCachePage(t, `<p id="p" style="display:block;width:80px;height:20px">x</p>`)
	screen.Handle(page.Handlers{
		Click: func(context.Context, page.Box) error {
			screen.Invalidate("p")
			return nil
		},
		ClickPaint: func(context.Context, page.Box) (bool, error) { return true, nil },
	})
	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	gen := screen.Generation()
	clickID(t, screen, "p")

	if got := screen.Generation(); got != gen {
		t.Fatalf("generation = %d, want %d", got, gen)
	}

	rect, ok := screen.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect after the paint-only click")
	}
	if !boxIn(rect, boxFor(t, screen, "p")) {
		t.Fatalf("rect %v misses the box", rect)
	}
}
