package login_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/login/login"
)

func newApp(t *testing.T, ctx context.Context) *login.App {
	t.Helper()

	app, err := login.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if len(app.PNG()) == 0 {
		t.Fatal("png is empty")
	}

	return app
}

func click(t *testing.T, ctx context.Context, app *login.App, id, action string) {
	t.Helper()

	x, y := center(t, app, id, action)
	if err := app.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}

func decodePNG(t *testing.T, app *login.App) image.Image {
	t.Helper()

	img, err := png.Decode(bytes.NewReader(app.PNG()))
	if err != nil {
		t.Fatal(err)
	}

	return img
}

func rgbaAt(img image.Image, x, y int) color.RGBA {
	r, g, b, a := img.At(x, y).RGBA()

	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)} //nolint:gosec // reason: image.RGBA returns 16-bit channels, so shifting right by 8 always fits a byte.
}
