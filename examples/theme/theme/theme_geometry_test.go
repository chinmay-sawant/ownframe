package theme_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/theme/theme"
)

// boxByID returns the box with id.
func boxByID(t *testing.T, app *theme.App, id string) ownframe.Box {
	t.Helper()

	for _, b := range app.Boxes() {
		if b.ID == id {
			return b
		}
	}

	t.Fatalf("no box id=%q", id)

	return ownframe.Box{}
}

// TestThemeSwitchKeepsGeometry pins the rule the toggle depends on: both
// themes set the same geometry, so a switch must not move or resize a box.
func TestThemeSwitchKeepsGeometry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := theme.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := boxByID(t, app, "toggle")
	if err := app.Click(ctx, before.X+before.W/2, before.Y+before.H/2); err != nil {
		t.Fatal(err)
	}

	after := boxByID(t, app, "toggle")
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("toggle box moved: %+v -> %+v", before, after)
	}
}
