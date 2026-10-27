package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// ticking reports a page with a frame callback. A Page marks mutated ops
// itself and stays on the buffer path.
type ticking interface {
	Ticking() bool
}

// partialState is the frame's retained pixels and the work that keeps them
// current. tiles is the content as a grid of replay tiles: a dirty rect
// repaints only the tiles it covers and the frame blits only damaged tiles.
// buf is the grid's retained-content handle (replay_tile_range.go).
//
// last is what the previous frame repainted, the grid's full rect on a
// rebuild or a blit and one rect per plan otherwise. plans, chrome, and whole
// are this frame's screen blits, kept for a reader that wants them; the
// tiled frame performs them itself. whole copies the content at scroll, the
// offset the last frame was blitted at. focus and screen are the last replay
// frame's window focus and screen size, so a focus regain or a resize
// repaints in full. live reports that the screen shows the page through the
// retained tiles at one to one, so a scaled, fitted, or bitmap frame forces
// the next replay frame to repaint everything. paints counts the tiles the
// frame repainted and paintPx their area; blits and blitPx count the tile
// blits; pins counts post-blit chrome draws. clipX is the dock clip.
type partialState struct {
	buf       *ebiten.Image
	tiles     *tileGrid
	gen       uint64
	last      []image.Rectangle
	plans     []repaintPlan
	chrome    []image.Rectangle
	blitRects []image.Rectangle
	mode      repaintMode
	paints    int
	paintPx   int
	blits     int
	blitPx    int
	pins      int
	focus     bool
	live      bool
	whole     bool
	scroll    image.Point
	screen    image.Point
	clipX     int
	blitOpts  ebiten.DrawImageOptions
}

// gridCurrent reports that the tile grid covers exactly this content size.
func (st partialState) gridCurrent(width, height int) bool {
	g := st.tiles

	return g != nil && g.width == width && g.height == height
}
