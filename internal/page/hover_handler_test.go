package page

import (
	"context"
	"errors"
	"testing"
)

func TestHoverHandlerFallbackAndError(t *testing.T) {
	p, err := New(Config{HTML: `<style>#x:hover{width:120px}</style><div id="x" style="height:20px"></div>`, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err = p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("paint failed")
	p.Handle(Handlers{Hover: func(context.Context, Box, Box) (bool, error) { return false, boom }})
	before := p.Generation()
	if err = p.Hover(ctx, 10, 10); !errors.Is(err, boom) {
		t.Fatalf("error = %v", err)
	}
	if p.HoverID() != "" || p.Generation() != before {
		t.Fatal("failed hover changed state")
	}
	p.Handle(Handlers{Hover: func(context.Context, Box, Box) (bool, error) { return false, nil }})
	if err = p.Hover(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if p.HoverID() != "x" || p.Generation() == before {
		t.Fatal("CSS fallback did not redraw")
	}
}
