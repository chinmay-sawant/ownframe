package drop_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/drop/drop"
)

func TestPickedPathPrints(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	path := filepath.Join(t.TempDir(), "picked.txt")

	app.Page().Testing().SetPicker(func(context.Context, string) (string, error) {
		return path, nil
	})

	click(t, ctx, app, "pick")

	if got := boxText(t, app, "paths"); !strings.Contains(got, path) {
		t.Fatalf("paths = %q", got)
	}
}

// click sends one click at the centre of the last box with that id.
func click(t *testing.T, ctx context.Context, app *drop.App, id string) {
	t.Helper()

	var x, y float64
	found := false

	for _, b := range app.Boxes() {
		if b.ID != id || b.W <= 0 || b.H <= 0 {
			continue
		}

		x, y = b.X+b.W/2, b.Y+b.H/2
		found = true
	}

	if !found {
		t.Fatalf("no box id=%q", id)
	}

	if err := app.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}
