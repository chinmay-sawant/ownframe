package window

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestDevDrawOverlaySmoke draws the overlay into an offscreen image in
// normal and stretched modes. The stroke hook records the computed screen
// rects, so the test finds each outline without reading pixels back.
func TestDevDrawOverlaySmoke(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.display = devTestDisplay()
	s.dev.ops = true
	s.devSetPicker(true)
	s.dev.haveHov = true
	s.dev.hovered = d.boxes[0]
	s.dev.havePin = true
	s.dev.pinned = d.boxes[0]

	var strokes []devRect
	s.dev.stroke = func(_ *ebiten.Image, r devRect, _ color.RGBA) {
		strokes = append(strokes, r)
	}

	img := ebiten.NewImage(1920, 1080)
	s.drawDevTools(img)

	box := d.boxes[0]
	want := s.devScreen(devRect{X: box.X, Y: box.Y, W: box.W, H: box.H})
	// The ops view strokes the display's two painted ops first, then the
	// hover and pin boxes.
	if len(strokes) != 4 || strokes[2] != want || strokes[3] != want {
		t.Fatalf("stretched outlines = %v, want the boxes %v last", strokes, want)
	}

	s.screenW, s.screenH = 800, 600
	s.drawDevTools(img)

	want = s.devScreen(devRect{X: box.X, Y: box.Y, W: box.W, H: box.H})
	if len(strokes) != 8 || strokes[6] != want || strokes[7] != want {
		t.Fatalf("normal outlines = %v, want the boxes %v last", strokes[4:], want)
	}

	// A page with no boxes still draws its dock on the right edge.
	d.boxes = nil
	s.dev.haveHov, s.dev.havePin = false, false
	s.drawDevTools(img)

	if s.dev.panel.W <= 0 || s.dev.panel.H <= 0 {
		t.Fatalf("panel = %+v, want a drawn panel", s.dev.panel)
	}

	if s.dev.panel.X+s.dev.panel.W != float64(s.screenW) || s.dev.panel.Y != 0 {
		t.Fatalf("panel = %+v, want the right dock", s.dev.panel)
	}
}
