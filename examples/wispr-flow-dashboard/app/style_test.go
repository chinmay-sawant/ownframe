package app

import (
	"context"
	"testing"
)

// TestStylePage checks the page root, the selected card, the per-app rows,
// and the display-list path.
func TestStylePage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("style")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-style"); !ok {
		t.Fatal("no page-style box")
	}

	if _, ok := findBox(app.Boxes(), "style-save"); !ok {
		t.Fatal("no style-save box")
	}

	cards := app.View().Style.Cards
	if len(cards) != 4 {
		t.Fatalf("style cards = %d", len(cards))
	}

	selected := 0
	for _, card := range cards {
		if card.Selected {
			selected++
		}
	}

	if selected != 1 || cards[1].Name != "Casual" || !cards[1].Selected {
		t.Fatalf("selected card = %+v", cards)
	}

	if len(app.View().Style.Apps) != 4 {
		t.Fatalf("per-app rows = %d", len(app.View().Style.Apps))
	}

	if app.Page().Display() == nil {
		t.Fatal("style fell back to the bitmap path")
	}
}

// TestStyleSaveLeavesNote checks the Save preference button reports.
func TestStyleSaveLeavesNote(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("style")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	save, ok := findBox(app.Boxes(), "style-save")
	if !ok {
		t.Fatal("no style-save box")
	}

	if save.Action != "style-save" {
		t.Fatalf("style-save action = %q", save.Action)
	}

	clickBox(t, app, save)

	if app.View().Note != "Style preference saved" {
		t.Fatalf("note = %q", app.View().Note)
	}
}
