package shapes_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/shapes/shapes"
)

func TestGalleryReplays(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	if !app.Replayable() {
		t.Fatal("gallery did not keep a display list")
	}

	if data := app.PNG(); len(data) == 0 {
		t.Fatal("no PNG")
	}

	before := app.Generation()

	// Resize so the redraw has a new frame to commit. An identical redraw
	// keeps the retained list, which is also a replayable frame.
	app.SetSize(shapes.DefaultWidth+10, shapes.DefaultHeight+10)
	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if after := app.Generation(); after <= before {
		t.Fatalf("generation = %d, want more than %d", after, before)
	}

	if !app.Replayable() {
		t.Fatal("gallery lost the display list after the redraw")
	}
}

func newApp(t *testing.T, ctx context.Context) *shapes.App {
	t.Helper()

	app, err := shapes.New()
	if err != nil {
		t.Fatal(err)
	}

	app.SetSize(shapes.DefaultWidth, shapes.DefaultHeight)

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	return app
}
