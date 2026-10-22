package png_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/png/png"
)

var pngSignature = []byte("\x89PNG\r\n\x1a\n")

func TestPNGCacheAndRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	first := app.PNG()
	if len(first) == 0 {
		t.Fatal("no PNG")
	}

	if !bytes.Equal(first, app.PNG()) {
		t.Fatal("second PNG call was not served from the cache")
	}

	click(t, ctx, app, "redraw")

	second := app.PNG()
	if bytes.Equal(first, second) {
		t.Fatal("the stamp did not change the picture")
	}

	path, err := app.SavePNG(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path) //nolint:gosec // reason: test reads back the file SavePNG just wrote under t.TempDir.
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.HasPrefix(data, pngSignature) {
		t.Fatalf("saved file does not start with the PNG signature")
	}
}

func newApp(t *testing.T, ctx context.Context) *png.App {
	t.Helper()

	app, err := png.New()
	if err != nil {
		t.Fatal(err)
	}

	app.SetSize(png.DefaultWidth, png.DefaultHeight)

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	return app
}

func click(t *testing.T, ctx context.Context, app *png.App, id string) {
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
