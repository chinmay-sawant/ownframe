package pagetest_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestClickAction(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML:   `<div id="n" data-action="add">1</div>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	var action string
	p.Handle(ownframe.Handlers{
		Click: func(_ context.Context, b ownframe.Box) error {
			action = b.Action
			return nil
		},
	})
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(context.Background(), p, "add"); err != nil {
		t.Fatal(err)
	}
	if action != "add" {
		t.Fatalf("action = %q", action)
	}
	_, ok := pagetest.BoxByID(p, "n")
	if !ok {
		t.Fatal("missing id n")
	}
}
