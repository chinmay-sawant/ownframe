package page

import (
	"context"
	"errors"
	"testing"
)

// patchPage builds a page that keeps the layout state an element patch needs.
func patchPage(t *testing.T, source string) *Page {
	t.Helper()

	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func mustPatch(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func mustRefuse(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrPatchRefused) {
		t.Fatalf("err = %v, want ErrPatchRefused", err)
	}
}
