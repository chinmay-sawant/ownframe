// Package replay draws a layout display list onto an Ebiten image. The window
// uses it to show the page as vector operations instead of one rasterized
// bitmap. The caller decides replayability first; a page it cannot draw keeps
// the bitmap path.
package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// Draw replays display in paint order. dx and dy move the page's top-left in
// pixels, so the window passes its scroll offset. A viewport-fixed op ignores
// the offset: the engine placed it against the viewport. Blend and isolation
// groups composite once through the paint plan.
func (renderer *Renderer) Draw(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	renderer.paintDisplay(dst, display, func(*layout.DisplayOp) (float64, float64, bool) {
		return dx, dy, true
	})
}

// DrawUnfixed replays only the operations that scroll with the page. A
// caller that paints a scrolled buffer draws the fixed ops separately, after
// the buffer, so they stay at the viewport.
func (renderer *Renderer) DrawUnfixed(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	renderer.paintDisplay(dst, display, func(op *layout.DisplayOp) (float64, float64, bool) {
		return dx, dy, !op.Fixed
	})
}

// DrawFixed replays only the viewport-fixed operations, untranslated.
func (renderer *Renderer) DrawFixed(dst *ebiten.Image, display *layout.Display) {
	renderer.paintDisplay(dst, display, func(op *layout.DisplayOp) (float64, float64, bool) {
		return 0, 0, op.Fixed
	})
}
