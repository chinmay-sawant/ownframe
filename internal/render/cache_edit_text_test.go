package render_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestCacheSetText checks the refusal cases and that the text lands in the
// cached tree.
func TestCacheSetText(t *testing.T) {
	ctx := context.Background()
	cache, err := render.NewCache(ctx, editSource, 320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}

	if cache.SetText("missing", "new") {
		t.Fatal("the edit on a missing element was not refused")
	}
	if cache.SetText("", "new") {
		t.Fatal("the edit without an id was not refused")
	}
	if !cache.SetText("b", "new") {
		t.Fatal("the edit was refused")
	}

	styled, err := cache.Relayout(ctx, 320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}
	if got := editText(t, styled); got != "new" {
		t.Fatalf("text %q, want new", got)
	}
}

// editText returns the text of the first text operation.
func editText(t *testing.T, styled *css.Document) string {
	t.Helper()

	display, err := render.DisplayListDocument(context.Background(), styled, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := range display.Ops {
		if display.Ops[i].Kind == render.OpText {
			return display.Ops[i].Text
		}
	}

	t.Fatal("no text operation")

	return ""
}
