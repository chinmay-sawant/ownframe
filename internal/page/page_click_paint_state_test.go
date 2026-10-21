package page

import (
	"context"
	"testing"
)

// TestClickPaintKeepsTheStateLayout checks that a paint-only click in a
// batch still lays out a state the sheets match: the handler painted the
// pixels, but the focus state needs a layout.
func TestClickPaintKeepsTheStateLayout(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#go{width:100px;height:40px}#go:focus{width:120px}</style>`+
		`<button id="go">Go</button>`)
	p.Handle(Handlers{
		ClickPaint: func(context.Context, Box) (bool, error) { return true, nil },
	})

	before := p.boxByID("go").W
	box := p.boxByID("go")
	x, y := box.X+box.W/2, box.Y+box.H/2

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

	if got := p.boxByID("go").W; got == before {
		t.Fatalf("width %v, want the focused width", got)
	}
}

// TestClickPaintSkipsTheLayoutOfAnUnusedState checks that a paint-only click
// in a batch on a page whose sheets match no state keeps the layout out.
func TestClickPaintSkipsTheLayoutOfAnUnusedState(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#go{width:100px;height:40px}</style><button id="go">Go</button>`)
	p.Handle(Handlers{
		ClickPaint: func(context.Context, Box) (bool, error) { return true, nil },
	})

	layouts := p.Stats().Layouts
	box := p.boxByID("go")
	x, y := box.X+box.W/2, box.Y+box.H/2

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

	if got := p.Stats().Layouts; got != layouts {
		t.Fatalf("layouts %d, want %d", got, layouts)
	}
}
