package telegram_test

import (
	"context"
	"math"
	"testing"
)

// closeEnough allows the one-pixel rounding a box may carry.
func closeEnough(got float64, want int) bool {
	return math.Abs(got-float64(want)) <= 1
}

func TestPinnedBarsFollowTheScroll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "", "open-anna")

	_, viewH := app.Page().Size()
	view := app.View()

	for _, off := range []int{0, 250, 800} {
		app.Page().Scroll().SetScrollOffset(0, off)

		if err := app.Redraw(ctx); err != nil {
			t.Fatal(err)
		}

		bar := boxByID(t, app, "threadbar")
		if want := off + view.InsetTop; !closeEnough(bar.Y, want) {
			t.Errorf("threadbar Y at offset %d = %.1f, want %d", off, bar.Y, want)
		}

		compose := boxByID(t, app, "composebar")
		want := off + viewH - view.InsetBottom - 63
		if !closeEnough(compose.Y, want) {
			t.Errorf("composebar Y at offset %d = %.1f, want %d", off, compose.Y, want)
		}
	}
}

func TestSheetLiftsTheComposer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "", "open-anna")
	click(t, ctx, app, "attach", "")

	if !app.View().AttachOpen {
		t.Fatal("attach sheet did not open")
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	off, _ := app.Page().Scroll().ScrollOffset()
	_, viewH := app.Page().Size()
	view := app.View()

	// The sheet stacks above the composer; the composer keeps the bottom.
	compose := boxByID(t, app, "composebar")
	want := off + viewH - view.InsetBottom - 63
	if !closeEnough(compose.Y, want) {
		t.Errorf("composebar Y with sheet at offset %d = %.1f, want %d", off, compose.Y, want)
	}

	camera := boxByID(t, app, "attach-camera")
	if camera.Y+camera.H > compose.Y+1 {
		t.Errorf("sheet row %.1f-%.1f is not above the composer %.1f", camera.Y, camera.Y+camera.H, compose.Y)
	}
}
