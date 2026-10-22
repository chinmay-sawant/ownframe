package page

import (
	"context"
	"testing"
)

func TestButtonsFocusAndDisable(t *testing.T) {
	p := bftPage(t, `<button id="a">A</button><button id="b" disabled>B</button><button id="c"><span>C</span></button>`)
	bftDraw(t, p, nil)
	ctx := context.Background()
	for _, want := range []string{"a", "c", "a"} {
		if err := p.FocusNext(ctx); err != nil {
			t.Fatal(err)
		}
		if p.FocusID() != want {
			t.Fatalf("focus %q, want %q", p.FocusID(), want)
		}
		if p.FocusedField() != "" {
			t.Fatal("button focus reported as an editable field")
		}
	}
	calls := 0
	p.Handle(Handlers{Click: func(context.Context, Box) error { calls++; return nil }})
	box := p.boxByID("b")
	if err := p.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("disabled button invoked handler")
	}
}

func TestButtonRouteAndDisabledRoute(t *testing.T) {
	p := bftPage(t, `<button id="off" disabled data-action="next">Off</button><button id="next" data-action="next"><span>Next</span></button>`)
	bftDraw(t, p, nil)
	if err := p.Route("next", `<p id="destination">Destination</p>`); err != nil {
		t.Fatal(err)
	}
	calls := 0
	p.Handle(Handlers{Click: func(context.Context, Box) error { calls++; return nil }})
	bftClick(t, p, "off")
	if p.boxByID("next").ID == "" || calls != 0 {
		t.Fatal("disabled button followed route or called handler")
	}
	bftClick(t, p, "next")
	if p.boxByID("destination").ID == "" || calls != 0 {
		t.Fatal("enabled button failed to route")
	}
}
