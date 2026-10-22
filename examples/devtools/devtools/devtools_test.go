package devtools_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/examples/devtools/devtools"
)

// TestDevToolsExampleHasEveryKind checks the example's reason to exist: one
// operation of every kind the replay draws.
func TestDevToolsExampleHasEveryKind(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := devtools.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	display := app.Display()
	if display == nil {
		t.Fatal("the example fell back to a bitmap")
	}

	kinds := map[layout.DisplayKind]int{}
	for i := range display.Ops {
		kinds[display.Ops[i].Kind]++
	}

	for _, want := range []layout.DisplayKind{
		layout.DisplayOpFillRect,
		layout.DisplayOpStrokeRect,
		layout.DisplayOpLine,
		layout.DisplayOpText,
		layout.DisplayOpImage,
		layout.DisplayOpGridRun,
	} {
		if kinds[want] == 0 {
			t.Fatalf("no operation of kind %d in the example", want)
		}
	}

	if len(app.Page().Testing().PNG()) == 0 {
		t.Fatal("PNG is empty")
	}
}

// TestDevToolsExampleCounter checks the click counter and the DevTools flag.
func TestDevToolsExampleCounter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := devtools.New()
	if err != nil {
		t.Fatal(err)
	}

	if !app.Page().Testing().DevTools() {
		t.Fatal("Config.DevTools did not start the overlay")
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "counter")

	if got := app.View().Count; got != 1 {
		t.Fatalf("count = %d, want 1", got)
	}
}

func click(t *testing.T, ctx context.Context, app *devtools.App, id string) {
	t.Helper()

	for _, b := range app.Boxes() {
		if b.ID != id || b.W <= 0 || b.H <= 0 {
			continue
		}

		if err := app.Click(ctx, b.X+b.W/2, b.Y+b.H/2); err != nil {
			t.Fatal(err)
		}

		return
	}

	t.Fatalf("no box id=%q", id)
}
