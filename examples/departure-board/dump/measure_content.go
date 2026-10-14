package dump

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// contentRect mirrors the page's contentRect and the window's contentSize:
// the canvas plus every box that overflows it. TakeDirtyRects clamps to this
// region, and a whole-content change returns it as one rect.
func contentRect(d *layout.Display, app *board.App) image.Rectangle {
	r := image.Rect(0, 0, d.Width, d.Height)
	for _, b := range app.Boxes() {
		r = r.Union(image.Rect(int(b.X), int(b.Y), int(b.X+b.W+0.5), int(b.Y+b.H+0.5)))
	}
	return r
}

// coversFull reports a rect equal to the whole content, the window's
// full-buffer trigger.
func coversFull(rects []image.Rectangle, content image.Rectangle) bool {
	for _, r := range rects {
		if r == content {
			return true
		}
	}
	return false
}
