package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestClickPaintEndsBatch checks that a paint-only click inside a redraw
// batch leaves no draw pending, so the batch waits for nothing.
func TestClickPaintEndsBatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<button id="go">Go</button>`)
	screen.Handle(page.Handlers{
		ClickPaint: func(context.Context, page.Box) (bool, error) { return true, nil },
	})

	before := screen.Stats()
	screen.BeginRedrawBatch()
	if err := screen.Focus(ctx, "go"); err != nil {
		t.Fatal(err)
	}
	clickID(t, screen, "go")
	if err := screen.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	after := screen.Stats()
	if after.Redraws != before.Redraws {
		t.Fatalf("redraws = %d, want %d", after.Redraws, before.Redraws)
	}
	if after.Layouts != before.Layouts {
		t.Fatalf("layouts = %d, want %d", after.Layouts, before.Layouts)
	}
	if got := screen.FocusID(); got != "go" {
		t.Fatalf("focus = %q, want go", got)
	}
}

// TestClickPaintRunsAfterClick checks that the paint handler runs after the
// click handler and after the control changed.
func TestClickPaintRunsAfterClick(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="c" type="checkbox">`)
	var order []string
	screen.Handle(page.Handlers{
		Click: func(context.Context, page.Box) error {
			order = append(order, "click")
			return nil
		},
		ClickPaint: func(context.Context, page.Box) (bool, error) {
			order = append(order, "paint")
			if !screen.FormChecked("c") {
				t.Error("the control did not change before the paint handler")
			}
			return true, nil
		},
	})

	clickID(t, screen, "c")

	if len(order) != 2 || order[0] != "click" || order[1] != "paint" {
		t.Fatalf("order = %v, want click then paint", order)
	}
}
