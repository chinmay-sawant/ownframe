package page

import (
	"context"
	"testing"
)

// fixedHeaderPage lays out a fixed header over a content block. The engine
// moves the fixed header behind the content box in the box list, so the hit
// test has to follow the paint order rather than the list order.
func fixedHeaderPage(t *testing.T) *Page {
	t.Helper()
	p, err := New(Config{
		HTML: `<div id="wrap"><div id="head" style="position:fixed;` +
			`width:100px;height:50px;background:#c00"></div></div>` +
			`<div id="body" style="width:100px;height:50px;background:#00c"></div>`,
		Width:  200,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestFixedBoxWinsOverContent checks that a fixed box wins the hit test
// against the content it covers, that a click lands on it, and that the
// content outside the fixed box still wins its own point.
func TestFixedBoxWinsOverContent(t *testing.T) {
	ctx := context.Background()
	p := fixedHeaderPage(t)

	spot := p.boxAt(10, 10)
	if !spot.ok || spot.id != "head" {
		t.Fatalf("boxAt = %+v, want the fixed header", spot)
	}

	var clicked Box
	p.Handle(Handlers{Click: func(_ context.Context, b Box) error {
		clicked = b
		return nil
	}})
	if err := p.Click(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if clicked.ID != "head" {
		t.Fatalf("clicked = %+v, want the fixed header", clicked)
	}

	spot = p.boxAt(10, 55)
	if !spot.ok || spot.id != "body" {
		t.Fatalf("boxAt = %+v, want the content block", spot)
	}
}
