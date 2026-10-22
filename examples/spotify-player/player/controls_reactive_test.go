package player

import (
	"testing"
)

func TestReactiveToggleState(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	for _, id := range []string{"play", "heart", "shuffle", "repeat"} {
		before := app.View()
		click(t, app, id)
		after := app.View()
		switch id {
		case "play":
			if after.Playing == before.Playing {
				t.Fatalf("%s did not toggle Playing", id)
			}
		case "heart":
			if after.Liked == before.Liked {
				t.Fatalf("%s did not toggle Liked", id)
			}
		case "shuffle":
			if after.Shuffle == before.Shuffle {
				t.Fatalf("%s did not toggle Shuffle", id)
			}
		case "repeat":
			if after.Repeat == before.Repeat {
				t.Fatalf("%s did not toggle Repeat", id)
			}
		}
	}

	redraw(t, app)
}

func TestReactiveTogglePartialDirty(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	for _, id := range []string{"play", "heart", "shuffle", "repeat"} {
		app.Page().TakeDirty()
		click(t, app, id)
		rect, ok := app.Page().TakeDirty()
		if !ok || rect.Empty() {
			t.Fatalf("%s left no dirty region", id)
		}
		box, found := findBox(app.Boxes(), id)
		if !found {
			t.Fatalf("no box %s", id)
		}
		cx := int(box.X + box.W/2)
		cy := int(box.Y + box.H/2)
		if cx < rect.Min.X || cx > rect.Max.X || cy < rect.Min.Y || cy > rect.Max.Y {
			t.Fatalf("%s center %d,%d outside dirty rect %v", id, cx, cy, rect)
		}
	}

	redraw(t, app)
}

func TestReactiveToggleStableIDs(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	for _, id := range []string{"play", "heart", "shuffle", "repeat"} {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("missing stable id %s", id)
		}
		if app.Page().Element(id).ID != id {
			t.Fatalf("retained element %s has wrong id", id)
		}
	}
}
