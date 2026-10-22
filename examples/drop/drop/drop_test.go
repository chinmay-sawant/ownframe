package drop_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/drop/drop"
)

func TestDroppedPathsPrint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	files := []ownframe.Drop{
		{Name: "photo.png", Path: "/home/me/photo.png"},
		{Name: "data.bin", Path: "/home/me/data.bin"},
		{Name: "docs", Path: "/home/me/docs", IsDir: true},
	}

	if err := app.Drop(ctx, files); err != nil {
		t.Fatal(err)
	}

	text := boxText(t, app, "paths")
	for _, want := range []string{"photo.png", "data.bin", "docs"} {
		if !strings.Contains(text, want) {
			t.Fatalf("paths = %q, missing %q", text, want)
		}
	}
	for _, leak := range []string{"/home/me/photo.png", "/home/me/data.bin", "/home/me/docs"} {
		if strings.Contains(text, leak) {
			t.Fatalf("paths = %q, leaks absolute %q", text, leak)
		}
	}
}

func TestBrowserFilePrintsName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	if err := app.Drop(ctx, []ownframe.Drop{{Name: "notes.txt"}}); err != nil {
		t.Fatal(err)
	}

	if text := boxText(t, app, "paths"); !strings.Contains(text, "notes.txt") {
		t.Fatalf("paths = %q", text)
	}
}

func newApp(t *testing.T, ctx context.Context) *drop.App {
	t.Helper()

	app, err := drop.New()
	if err != nil {
		t.Fatal(err)
	}

	app.SetSize(drop.DefaultWidth, drop.DefaultHeight)

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	return app
}

func boxText(t *testing.T, app *drop.App, id string) string {
	t.Helper()

	for _, box := range app.Boxes() {
		if box.ID == id {
			return box.Text
		}
	}

	t.Fatalf("no box id=%q", id)

	return ""
}
