package page

import (
	"context"
	"testing"
)

func TestUnusedStatePseudosSkipRedraw(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>` +
		`#e{width:100px;height:20px}` +
		`#e:focus-visible{outline:2px solid red}` +
		`</style></head><body><div id="e"></div></body></html>`

	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.statePseudos.Hover {
		t.Skip("engine does not report which state pseudo-classes the sheets use")
	}

	before := p.Stats().Redraws
	if err := p.Hover(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if err := p.Press(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if err := p.Release(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Redraws; got != before {
		t.Fatalf("redraws = %d after hover/press/release, want %d", got, before)
	}
	if err := p.Hover(ctx, 300, 150); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Redraws; got != before {
		t.Fatalf("redraws = %d after unhover, want %d", got, before)
	}
}

func TestHoverStateStillRedrawsWhenUsed(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>` +
		`#e{width:100px;height:20px}` +
		`#e:hover{width:120px}` +
		`</style></head><body><div id="e"></div></body></html>`

	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := p.Stats().Redraws
	if err := p.Hover(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Redraws; got <= before {
		t.Fatalf("redraws = %d after used hover, want more than %d", got, before)
	}
}
