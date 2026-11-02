package page

import (
	"context"
	"testing"
)

func TestKeyboardButtonActivationAndFormSubmit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := New(Config{HTML: `<button id="action" type="button" data-action="run">Run</button><input id="field">`, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}
	clicks, submits := 0, 0
	p.Handle(Handlers{Click: func(_ context.Context, box Box) error {
		if box.Action == "run" {
			clicks++
		}
		return nil
	}, Submit: func(context.Context) error { submits++; return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "action"); err != nil {
		t.Fatal(err)
	}
	if err := p.KeyDown(ctx, "space"); err != nil {
		t.Fatal(err)
	}
	if clicks != 1 || submits != 0 {
		t.Fatalf("Space: clicks=%d submits=%d", clicks, submits)
	}
	if err := p.KeyDown(ctx, "enter"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if clicks != 2 || submits != 0 {
		t.Fatalf("Enter: clicks=%d submits=%d", clicks, submits)
	}
	if err := p.Focus(ctx, "field"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if clicks != 2 || submits != 1 {
		t.Fatalf("field submit: clicks=%d submits=%d", clicks, submits)
	}
}

func TestKeyboardButtonFollowsRoute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := New(Config{HTML: `<button id="next" data-action="next">Next</button>`, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Route("next", `<p>Next page</p>`); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "next"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != `<p>Next page</p>` {
		t.Fatal("keyboard activation did not navigate")
	}
}

func TestDisabledFocusedButtonDoesNotSubmit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := New(Config{HTML: `<button id="action" type="button">Action</button>`, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}
	submits := 0
	p.Handle(Handlers{Submit: func(context.Context) error { submits++; return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "action"); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceHTML(ctx, p, `<button id="action" type="button" disabled>Action</button>`); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if submits != 0 {
		t.Fatal("disabled button submitted")
	}
}

func TestExplicitSubmitButtonUsesSubmitHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := New(Config{HTML: `<button id="submit" type="submit">Submit</button>`, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}
	submits := 0
	p.Handle(Handlers{Submit: func(context.Context) error { submits++; return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "submit"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if submits != 1 {
		t.Fatalf("submits=%d", submits)
	}
}
