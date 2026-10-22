package login_test

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/login/login"
)

var (
	pageBackground = color.RGBA{R: 0xf4, G: 0xf1, B: 0xea, A: 0xff}
	cardBackground = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	blueButton     = color.RGBA{R: 0x1a, G: 0x56, B: 0xdb, A: 0xff}
	disabledButton = color.RGBA{R: 0xe3, G: 0xde, B: 0xd4, A: 0xff}
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

// center uses the last matching box so a nested element wins over its parent.
func center(t *testing.T, app *login.App, id, action string) (float64, float64) {
	t.Helper()

	var x, y float64
	found := false

	for _, b := range app.Boxes() {
		if id != "" && b.ID != id {
			continue
		}

		if action != "" && b.Action != action {
			continue
		}

		if b.W <= 0 || b.H <= 0 {
			continue
		}

		x = b.X + b.W/2
		y = b.Y + b.H/2
		found = true
	}

	if !found {
		t.Fatalf("no box id=%q action=%q:%s", id, action, dumpBoxes(app))
	}

	return x, y
}

func dumpBoxes(app *login.App) string {
	var b strings.Builder

	for _, box := range app.Boxes() {
		fmt.Fprintf(
			&b,
			"\n%s id=%q action=%q text=%q xywh=%.1f,%.1f,%.1f,%.1f",
			box.Tag,
			box.ID,
			box.Action,
			box.Text,
			box.X,
			box.Y,
			box.W,
			box.H,
		)
	}

	if b.Len() == 0 {
		return "\n(no boxes)"
	}

	return b.String()
}

func boxByID(t *testing.T, app *login.App, id string) ownframe.Box {
	t.Helper()

	for _, box := range app.Boxes() {
		if box.ID == id {
			return box
		}
	}

	t.Fatalf("no box id=%q:%s", id, dumpBoxes(app))

	return ownframe.Box{}
}

func boxByTag(t *testing.T, app *login.App, tag string) ownframe.Box {
	t.Helper()

	for _, box := range app.Boxes() {
		if box.Tag == tag {
			return box
		}
	}

	t.Fatalf("no box tag=%q:%s", tag, dumpBoxes(app))

	return ownframe.Box{}
}

func boxText(t *testing.T, app *login.App, id string) string {
	t.Helper()

	var text string
	found := false

	for _, b := range app.Boxes() {
		if b.ID != id {
			continue
		}

		text = b.Text
		found = true
	}

	if !found {
		t.Fatalf("no box id=%q:%s", id, dumpBoxes(app))
	}

	return text
}

// buttonFill samples the flat fill inside the button edge, away from the label.
func buttonFill(t *testing.T, app *login.App) color.RGBA {
	t.Helper()

	btn := boxByID(t, app, "login")
	img := decodePNG(t, app)

	return rgbaAt(img, int(btn.X)+4, int(btn.Y+btn.H/2))
}
