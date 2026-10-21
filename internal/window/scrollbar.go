package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	scrollbarThickness = 8
	scrollbarMinThumb  = 24
)

const scrollThumbHex = 0xbdbdbdff

var scrollThumb = devRGBA(scrollThumbHex)

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
	viewW, viewH := s.viewSize()
	thick := float32(scrollbarThickness)

	if scrollbarVisible(contentH, viewH) {
		pos, length := scrollbarThumb(viewH, contentH, viewH, s.scrollY)
		x := float32(viewW) - thick
		vector.FillRect(dst, x, pos, thick, length, scrollThumb, false)
		s.frameNote(frameRect(float64(x), float64(pos), float64(thick), float64(length)))
	}

	if scrollbarVisible(contentW, viewW) {
		pos, length := scrollbarThumb(viewW, contentW, viewW, s.scrollX)
		y := float32(viewH) - thick
		vector.FillRect(dst, pos, y, length, thick, scrollThumb, false)
		s.frameNote(frameRect(float64(pos), float64(y), float64(length), float64(thick)))
	}
}
