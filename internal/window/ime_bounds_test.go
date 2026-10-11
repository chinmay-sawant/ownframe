package window

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestIMEBoundsMatchZoomedScroll(t *testing.T) {
	app := &fakeScreen{width: 200, height: 200}
	s := &shell{app: app, screenW: 200, screenH: 200, scrollX: 20, scrollY: 40, pageZoom: 2}
	box := layout.Box{X: 10, Y: 20, W: 80, H: 20}
	if got := s.imeScreenRect(box); got != image.Rect(40, 80, 200, 120) {
		t.Fatal(got)
	}
}

func TestIMEBoundsMatchPendingResize(t *testing.T) {
	app := &fakeScreen{width: 200, height: 200}
	s := &shell{app: app, screenW: 400, screenH: 100}
	box := layout.Box{X: 10, Y: 20, W: 80, H: 20}
	if got := s.imeScreenRect(box); got != image.Rect(20, 10, 180, 20) {
		t.Fatal(got)
	}
}
