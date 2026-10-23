package page_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestLoadBackForward(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := `<div id="a">A</div>`
	second := `<div id="b">B</div>`
	screen := navPage(t, first)
	navAt(t, screen, first, "a")

	if err := screen.Load(ctx, second); err != nil {
		t.Fatal(err)
	}

	navAt(t, screen, second, "b")
	if err := screen.Back(ctx); err != nil {
		t.Fatal(err)
	}

	navAt(t, screen, first, "a")
	if err := screen.Forward(ctx); err != nil {
		t.Fatal(err)
	}

	navAt(t, screen, second, "b")
	if err := screen.Forward(ctx); !errors.Is(err, page.ErrNoHistory) {
		t.Fatalf("forward = %v", err)
	}

	if err := screen.Back(ctx); err != nil {
		t.Fatal(err)
	}

	if err := screen.Back(ctx); !errors.Is(err, page.ErrNoHistory) {
		t.Fatalf("back = %v", err)
	}
}

func TestRouteClickLoadsHTML(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	next := `<div id="b">B</div>`
	screen := navPage(t, `<div id="a" data-action="go">A</div>`)
	if err := screen.Route("go", `<div id="c">C</div>`); err != nil {
		t.Fatal(err)
	}
	if err := screen.Route("go", next); err != nil {
		t.Fatal(err)
	}

	x, y := boxCenter(t, screen, "a")
	if err := screen.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	navAt(t, screen, next, "b")
}

func TestLoadRejectsBlankHTML(t *testing.T) {
	t.Parallel()

	screen := navPage(t, `<div id="a">A</div>`)
	err := screen.Load(context.Background(), " ")
	if !errors.Is(err, page.ErrEmptyHTML) {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadAfterBackReplacesForwardHistory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	first := `<div id="a">A</div>`
	second := `<div id="b">B</div>`
	third := `<div id="c">C</div>`
	screen := navPage(t, first)
	if err := screen.Load(ctx, second); err != nil {
		t.Fatal(err)
	}
	if err := screen.Back(ctx); err != nil {
		t.Fatal(err)
	}
	if err := screen.Load(ctx, third); err != nil {
		t.Fatal(err)
	}
	if err := screen.Back(ctx); err != nil {
		t.Fatal(err)
	}
	navAt(t, screen, first, "a")
	if err := screen.Forward(ctx); err != nil {
		t.Fatal(err)
	}
	navAt(t, screen, third, "c")
	if err := screen.Forward(ctx); !errors.Is(err, page.ErrNoHistory) {
		t.Fatalf("forward after replacement = %v", err)
	}
}
