package render_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

const editSource = `<style>#b{width:100px;height:20px}</style><div id="b">old</div>`

// TestCacheSetAttribute checks the refusal cases and that the edit lands in
// the cached tree.
func TestCacheSetAttribute(t *testing.T) {
	ctx := context.Background()
	cache, err := render.NewCache(ctx, editSource, 320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}

	if cache.SetAttribute("missing", "class", "on") {
		t.Fatal("the edit on a missing element was not refused")
	}
	if cache.SetAttribute("", "class", "on") {
		t.Fatal("the edit without an id was not refused")
	}
	if cache.SetAttribute("b", "", "on") {
		t.Fatal("the edit without a name was not refused")
	}
	if !cache.SetAttribute("b", "style", "width:150px") {
		t.Fatal("the edit was refused")
	}

	styled, err := cache.Relayout(ctx, 320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}
	if got := editWidth(t, styled); got != 150 {
		t.Fatalf("width %v, want 150", got)
	}
}

// editWidth returns the width of the box with the id b.
func editWidth(t *testing.T, styled *css.Document) float64 {
	t.Helper()

	display, err := render.DisplayListDocument(context.Background(), styled, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, box := range display.Boxes {
		if box.ID == "b" {
			return box.W
		}
	}

	t.Fatal("no box b")

	return 0
}
