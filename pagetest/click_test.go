package pagetest_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestClickFindsBox(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML:   `<button id="go" data-action="add">n</button>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	var hit string
	p.Handle(ownframe.Handlers{
		Click: func(_ context.Context, b ownframe.Box) error {
			hit = b.ID
			return nil
		},
	})
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(context.Background(), p, "go"); err != nil {
		t.Fatal(err)
	}
	if hit != "go" {
		t.Fatalf("hit = %q", hit)
	}
	if err := pagetest.Click(context.Background(), p, "nope"); !errors.Is(err, pagetest.ErrNoBox) {
		t.Fatalf("err = %v", err)
	}
}

func TestTypeEditsFocusedField(t *testing.T) {
	ctx := context.Background()
	p, err := ownframe.New(ownframe.Config{
		HTML: `<input id="name" value="A">`, Width: 320, Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Form().Focus(ctx, "name"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Type(ctx, p, "da"); err != nil {
		t.Fatal(err)
	}
	if got := p.Form().FormValue("name"); got != "Ada" {
		t.Fatalf("typed value = %q", got)
	}
	if err := pagetest.Type(ctx, nil, "x"); !errors.Is(err, pagetest.ErrNoBox) {
		t.Fatalf("nil page = %v", err)
	}
}
