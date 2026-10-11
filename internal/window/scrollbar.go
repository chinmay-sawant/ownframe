package window

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	scrollbarThickness = 8
	scrollbarMinThumb  = 24
)

var scrollThumb = color.RGBA{R: 0xBD, G: 0xBD, B: 0xBD, A: 0xFF}

func (s *shell) canvasSize() (int, int) {
	if s.display != nil {
		return s.display.Width, s.display.Height
	}

	if s.img != nil {
		b := s.img.Bounds()

		return b.Dx(), b.Dy()
	}

	return s.app.Size()
}

// scrollbarVisible reports whether content overflows the viewport.
func scrollbarVisible(content, viewport int) bool {
	return content > viewport && viewport > 0
}

// drawScrollbars overlays the thumbs a page larger than the window needs.
// Only the thumb is painted, so the page background stays visible.
func (s *shell) drawScrollbars(dst *ebiten.Image) {
	if !s.scrollbarsEnabled() {
		return
	}

	contentW, contentH := s.contentSize()
	thick := float32(scrollbarThickness)

	if scrollbarVisible(contentH, s.screenH) {
		pos, length := scrollbarThumb(s.screenH, contentH, s.screenH, s.scrollY)
		x := float32(s.screenW) - thick
		vector.FillRect(dst, x, pos, thick, length, scrollThumb, false)
	}

	if scrollbarVisible(contentW, s.screenW) {
		pos, length := scrollbarThumb(s.screenW, contentW, s.screenW, s.scrollX)
		y := float32(s.screenH) - thick
		vector.FillRect(dst, pos, y, length, thick, scrollThumb, false)
	}
}
