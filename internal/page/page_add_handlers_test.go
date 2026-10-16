package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestAddHandlersKeepsPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := page.New(page.Config{HTML: `<div>hi</div>`, Width: 200, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	var down, up int
	p.Handle(page.Handlers{
		KeyDown: func(context.Context, string) error { down++; return nil },
	})
	p.AddHandlers(page.Handlers{
		KeyUp: func(context.Context, string) error { up++; return nil },
	})
	if err := p.KeyDown(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if err := p.KeyUp(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if down != 1 || up != 1 {
		t.Fatalf("down=%d up=%d", down, up)
	}
	p.AddHandlers(page.Handlers{})
	if err := p.KeyDown(ctx, "b"); err != nil {
		t.Fatal(err)
	}
	if down != 2 {
		t.Fatalf("down=%d", down)
	}
}
