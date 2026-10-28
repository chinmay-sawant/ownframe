package main

import (
	"context"
	"testing"
)

func TestEditUndoRedoHistory(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	enterEditAtFirstLine(t, app)

	initial := app.edit.Text()
	edit := app.Page().Edit()
	for _, text := range []string{"a", "b"} {
		if err := edit.Type(ctx, text); err != nil {
			t.Fatal(err)
		}
	}

	assertText := func(want string) {
		t.Helper()
		if got := app.edit.Text(); got != want {
			t.Fatalf("edit text = %q, want %q", got, want)
		}
	}
	assertText("ab" + initial)

	for range 2 {
		if err := edit.Undo(ctx); err != nil {
			t.Fatal(err)
		}
	}
	assertText(initial)
	if err := edit.Redo(ctx); err != nil {
		t.Fatal(err)
	}
	assertText("a" + initial)
	if err := edit.Redo(ctx); err != nil {
		t.Fatal(err)
	}
	assertText("ab" + initial)

	if err := edit.Undo(ctx); err != nil {
		t.Fatal(err)
	}
	if err := edit.Type(ctx, "c"); err != nil {
		t.Fatal(err)
	}
	if err := edit.Redo(ctx); err != nil {
		t.Fatal(err)
	}
	assertText("ac" + initial)
}
