package app

import (
	"context"
	"testing"
)

// TestDictionaryPage checks the word rows and the replay path.
func TestDictionaryPage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("dictionary")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-dictionary"); !ok {
		t.Fatal("no dictionary page")
	}

	if _, ok := findBox(app.Boxes(), "word-row-6"); !ok {
		t.Fatal("no word-row-6 box")
	}

	addWord, ok := findBox(app.Boxes(), "word-add")
	if !ok {
		t.Fatal("no word-add box")
	}

	if addWord.Action != "word-add" {
		t.Fatalf("word-add action = %q", addWord.Action)
	}

	if got := len(app.View().Dictionary.Words); got != 6 {
		t.Fatalf("words = %d", got)
	}

	if app.Page().Testing().Display() == nil {
		t.Fatal("dictionary fell back to a bitmap")
	}
}
