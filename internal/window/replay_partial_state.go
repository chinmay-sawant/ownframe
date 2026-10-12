package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// ticking reports a page with a frame callback. A tick changes operations in
// place without a dirty rect, so such a page keeps the full replay.
type ticking interface {
	Ticking() bool
}

// partialState is the persistent replay buffer and the frame it holds. last is
// what the previous frame repainted, the buffer's full rect on a rebuild or a
// blit and one rect per plan otherwise. plans, chrome, and whole are this
// frame's screen blits, kept until the caller draws them: whole copies the
// buffer at scroll, which is the offset the last frame was blitted at. focus
// and screen are the last replay frame's window focus and screen size, so a
// focus regain or a resize repaints in full. live reports that the screen
// shows the page through the buffer at one to one, so a scaled, fitted, or
// bitmap frame forces the next replay frame to repaint everything.
type partialState struct {
	buf    *ebiten.Image
	gen    uint64
	width  int
	height int
	last   []image.Rectangle
	plans  []repaintPlan
	chrome []image.Rectangle
	mode   repaintMode
	focus  bool
	live   bool
	whole  bool
	scroll image.Point
	screen image.Point
}
