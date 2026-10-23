package page

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/hit"
)

func overlapFixture() []Box {
	return []Box{
		{ID: "outer", X: 0, Y: 0, W: 100, H: 100},
		{ID: "inner", X: 10, Y: 10, W: 20, H: 20},
		{ID: "", X: 12, Y: 12, W: 2, H: 2},
	}
}

func overlapPage(t *testing.T) *Page {
	t.Helper()
	p, err := New(Config{HTML: `<div id="outer"><div id="inner"></div></div>`, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestBoxAtUnifiesPointerWinners pins one overlapping fixture where hit,
// boxIDAt, and boxAt agree, then drives Click, Press, and Hover at the same
// point and asserts all three land on the innermost id.
func TestBoxAtUnifiesPointerWinners(t *testing.T) {
	ctx := context.Background()
	p := overlapPage(t)
	p.setBoxes(overlapFixture())

	box, ok := hit.Hit(p.Boxes(), 13, 13)
	if !ok || box.ID != "" {
		t.Fatalf("hit = %+v, %v, want the innermost anonymous box", box, ok)
	}
	if id := p.boxIDAt(13, 13); id != "inner" {
		t.Fatalf("boxIDAt = %q, want inner", id)
	}
	if spot := p.boxAt(13, 13); !spot.ok || spot.id != "inner" {
		t.Fatalf("boxAt = %+v, want inner", spot)
	}

	p.setBoxes(overlapFixture())
	if err := p.Press(ctx, 13, 13); err != nil {
		t.Fatal(err)
	}
	if p.activeID() != "inner" {
		t.Fatalf("press = %q, want inner", p.activeID())
	}

	p.setBoxes(overlapFixture())
	if err := p.Hover(ctx, 13, 13); err != nil {
		t.Fatal(err)
	}
	if p.hoverID() != "inner" {
		t.Fatalf("hover = %q, want inner", p.hoverID())
	}

	p.setBoxes(overlapFixture())
	var clicked Box
	p.Handle(Handlers{Click: func(_ context.Context, b Box) error {
		clicked = b
		return nil
	}})
	if err := p.Click(ctx, 13, 13); err != nil {
		t.Fatal(err)
	}
	if clicked.ID != "inner" {
		t.Fatalf("clicked = %+v, want inner", clicked)
	}
}
