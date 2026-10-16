package page

import (
	"context"
	"testing"
)

func frontPage(t *testing.T, html string) *Page {
	t.Helper()
	p, err := New(Config{HTML: html, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestSelectFrontThreeStates pins the preference all sinks call: display
// wins, image only reports its bounds and fallback, and an undrawn page
// falls back to the frame size.
func TestSelectFrontThreeStates(t *testing.T) {
	front := SelectFront(frontPage(t, `<p>x</p>`))
	if front.Display != nil || front.Image != nil || front.Fallback {
		t.Fatalf("idle = %+v, want neither", front)
	}
	if front.Width != 120 || front.Height != 80 {
		t.Fatalf("idle size = %d x %d, want 120 x 80", front.Width, front.Height)
	}

	drawn := frontPage(t, `<p>hi</p>`)
	if err := drawn.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	front = SelectFront(drawn)
	if front.Display == nil || front.Display != drawn.Display() {
		t.Fatalf("display = %+v, want the page display", front.Display)
	}
	if front.Fallback || front.Image != nil || front.Generation != drawn.Generation() {
		t.Fatalf("display page = %+v, want no fallback", front)
	}
	if d := drawn.Display(); front.Width != d.Width || front.Height != d.Height {
		t.Fatalf("display size = %d x %d", front.Width, front.Height)
	}

	fb := frontPage(t, `<style>div{width:60px;height:40px;background:#123456;`+
		`border-radius:50% / 20%;}</style><div>x</div>`)
	if err := fb.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fb.Display() != nil || fb.Image() == nil {
		t.Fatal("elliptical border did not fall back to a bitmap")
	}
	front = SelectFront(fb)
	if front.Display != nil || !front.Fallback || front.Image == nil {
		t.Fatalf("fallback = %+v, want image only", front)
	}
	b := fb.Image().Bounds()
	if front.Width != b.Dx() || front.Height != b.Dy() {
		t.Fatalf("image size = %d x %d, want %d x %d", front.Width, front.Height, b.Dx(), b.Dy())
	}
}
