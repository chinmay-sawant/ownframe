package player

import "testing"

func TestShelfAndSidebarSpacing(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
	boxes := app.Boxes()
	card, ok := findBox(boxes, "card-0")
	if !ok {
		t.Fatal("Recently played card is missing")
	}
	titleY := -1.0
	for _, box := range boxes {
		if box.Tag == "span" && box.Text == app.View().Shelf[0].Title {
			titleY = box.Y
			break
		}
	}
	if titleY < card.Y+148 {
		t.Fatalf("shelf title starts too close to cover: offset %.1f px", titleY-card.Y)
	}
	first, _ := findBox(boxes, "list-0")
	second, _ := findBox(boxes, "list-1")
	if gap := second.Y - (first.Y + first.H); gap < 6 {
		t.Fatalf("sidebar library row gap = %.1f px", gap)
	}
}
