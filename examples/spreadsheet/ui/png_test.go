package ui

import (
	"bytes"
	"context"
	"image/png"
	"testing"
)

// pixel decodes the last PNG and returns one pixel.
func pixel(t *testing.T, app *App, x, y int) [3]uint8 {
	t.Helper()

	data := app.page.Testing().PNG()
	if len(data) == 0 {
		t.Fatal("no PNG")
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	b := img.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		t.Fatalf("pixel %d,%d outside %v", x, y, b)
	}

	r, g, bb, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()

	return [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(bb >> 8)} //nolint:gosec // reason: RGBA>>8 yields 0-255 which always fits in uint8
}

// near reports whether two colors are within a tolerance per channel.
func near(a, b [3]uint8) bool {
	for i := range a {
		d := int(a[i]) - int(b[i])
		if d < -8 || d > 8 {
			return false
		}
	}

	return true
}

func TestFrozenChromeAndScrollingCells(t *testing.T) {
	app := newTestApp(t, newFake())
	settle(t, app)

	// A1:B2 selected, so the selection rectangle marks content coordinates.
	app.selectRect(0, 0, 1, 1)
	if err := app.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	button := [3]uint8{255, 255, 255}
	sel := [3]uint8{0xdc, 0xe8, 0xff}
	if got := pixel(t, app, 9, 7); !near(got, button) {
		t.Fatalf("toolbar at scroll 0 = %v", got)
	}

	if got := pixel(t, app, HeadW+6, ChromeH+6); !near(got, sel) {
		t.Fatalf("selection at scroll 0 = %v", got)
	}

	// Scroll far down: the chrome stays, the cells move away.
	app.page.SetScrollOffset(0, 1000)
	app.page.Testing().StepScrollWindow()
	if err := app.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := pixel(t, app, 9, 7); !near(got, button) {
		t.Fatalf("toolbar at scroll 1000 = %v", got)
	}

	if got := pixel(t, app, HeadW+6, ChromeH+6); near(got, sel) {
		t.Fatalf("selection did not scroll away: %v", got)
	}
}
