package page_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestDropErrorSkipsRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{HTML: `<p>hi</p>`, Width: 200, Height: 120})
	if err != nil {
		t.Fatal(err)
	}

	want := errors.New("drop refused")
	screen.Handle(page.Handlers{
		Drop: func(context.Context, []page.Drop) error { return want },
	})

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Generation()

	if err := screen.Drop(ctx, []page.Drop{{Name: "a.txt"}}); !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}

	if screen.Generation() != before {
		t.Fatalf("generation = %d, before = %d", screen.Generation(), before)
	}
}

func TestDropRedrawsWithoutHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{HTML: `<p>hi</p>`, Width: 200, Height: 120})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Generation()

	if err := screen.Drop(ctx, []page.Drop{{Name: "a.txt"}}); err != nil {
		t.Fatal(err)
	}

	// No handler changed page input, so the redraw runs but commits an
	// identical frame: one more redraw, no new generation.
	if got := screen.Stats().Redraws; got != 2 {
		t.Fatalf("redraws = %d, want 2", got)
	}
	if screen.Generation() != before {
		t.Fatalf("generation = %d, before = %d", screen.Generation(), before)
	}
}
