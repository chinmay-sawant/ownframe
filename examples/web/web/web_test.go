package web_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/web/web"
)

func TestWebCounterAndNote(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := web.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if len(app.Page().PNG()) == 0 {
		t.Fatal("PNG is empty after Redraw")
	}

	click(t, ctx, app, "inc")
	click(t, ctx, app, "inc")

	if got := app.View().Count; got != 2 {
		t.Fatalf("Count = %d, want 2", got)
	}

	click(t, ctx, app, "note")
	if err := app.Type(ctx, "hello"); err != nil {
		t.Fatal(err)
	}

	if got := app.FormValue("note"); got != "hello" {
		t.Fatalf("FormValue(note) = %q, want hello", got)
	}

	if got := app.View().Note; got != "hello" {
		t.Fatalf("View Note = %q, want hello", got)
	}

	click(t, ctx, app, "reset")
	if got := app.View().Count; got != 0 {
		t.Fatalf("Count after reset = %d, want 0", got)
	}
}

func click(t *testing.T, ctx context.Context, app *web.App, id string) {
	t.Helper()

	var x, y float64
	found := false

	for _, b := range app.Boxes() {
		if b.ID != id || b.W <= 0 || b.H <= 0 {
			continue
		}

		x = b.X + b.W/2
		y = b.Y + b.H/2
		found = true
	}

	if !found {
		t.Fatalf("no box id=%q", id)
	}

	if err := app.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}
