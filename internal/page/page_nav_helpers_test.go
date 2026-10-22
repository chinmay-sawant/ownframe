package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func navPage(t *testing.T, html string) *page.Page {
	t.Helper()

	screen, err := page.New(page.Config{
		HTML:   html,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return screen
}

func navAt(t *testing.T, screen *page.Page, html, id string) {
	t.Helper()

	if screen.HTML() != html {
		t.Fatalf("html = %q", screen.HTML())
	}

	boxCenter(t, screen, id)
}
