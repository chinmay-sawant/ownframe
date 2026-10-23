package draw

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestRedrawKeySkipsUnmatchedState(t *testing.T) {
	t.Parallel()

	view := Viewport{Width: 10, Height: 20}
	state := render.State{Hover: "a", Active: "b", Focus: "c"}
	got := KeyFor("src", view, state, render.StatePseudos{Focus: true}, 3)
	if got.Hover != "" || got.Active != "" || got.Focus != "c" {
		t.Fatalf("state %+v", got)
	}
	if got.Source != "src" || got.Width != 10 || got.Height != 20 || got.Images != 3 {
		t.Fatalf("frame %+v", got)
	}
	if state.Hover != "a" {
		t.Fatal("key mutated the caller")
	}
	if !Unchanged(true, got, got) || Unchanged(false, got, got) {
		t.Fatal("unchanged")
	}
}
