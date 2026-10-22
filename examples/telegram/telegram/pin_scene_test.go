package telegram_test

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/telegram/telegram"
)

// shot renders the page as the window would show it after scrolling to off:
// Pin runs with pinOff, and a body translate brings document rows
// [off, off+viewH) into the viewport. pinOff = off pins the bars; pinOff =
// off-viewH pushes them above the viewport, exposing the content underneath.
func shot(t *testing.T, app *telegram.App, ctx context.Context, off, pinOff int) *image.RGBA {
	t.Helper()

	theme := fmt.Sprintf("body { transform: translateY(-%dpx); }", off)
	if err := app.Page().SetTheme(theme); err != nil {
		t.Fatal(err)
	}

	app.Page().Scroll().SetScrollOffset(0, pinOff)

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	return toRGBA(t, app.PNG())
}

// newShotApp opens a chat with four photo messages so the thread scrolls.
func newShotApp(t *testing.T, ctx context.Context) *telegram.App {
	t.Helper()

	app := newApp(t, ctx)
	click(t, ctx, app, "", "open-anna")

	for i := 0; i < 4; i++ {
		click(t, ctx, app, "attach", "")
		app.RequestAttach("gallery")
		app.QueuePhoto(solidPNG(t, 400, 640, color.RGBA{R: 200, G: 40, B: 40, A: 255}))

		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}
	}

	return app
}
