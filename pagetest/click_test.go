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
