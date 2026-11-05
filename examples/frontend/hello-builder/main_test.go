package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestGreetUpdatesMessage(t *testing.T) {
	ctx := context.Background()
	page, state, err := newPage()
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if state.Message != "Hello, world!" {
		t.Fatalf("message = %q", state.Message)
	}
	if !strings.Contains(page.HTML(), "Hello, world!") {
		t.Fatal("opening message was not rendered")
	}
	if _, ok := pagetest.BoxByID(page, "greet"); !ok {
		t.Fatal("greet was not rendered")
	}
	if err := pagetest.Click(ctx, page, "greet"); err != nil {
		t.Fatal(err)
	}
	if state.Message != "Hello from Go!" {
		t.Fatalf("message = %q", state.Message)
	}
	if !strings.Contains(page.HTML(), "Hello from Go!") || strings.Contains(page.HTML(), "Hello, world!") {
		t.Fatalf("tree was not rebuilt: %s", page.HTML())
	}
}
