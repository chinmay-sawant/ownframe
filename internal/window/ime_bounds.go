package window

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// imeScreenRect maps the already scroll-adjusted control box to window pixels.
// Density conversion belongs to EbitenView, so it is not applied here.
func (s *shell) imeScreenRect(box layout.Box) image.Rectangle {
	sx, sy, ox, oy := s.zoom(), s.zoom(), 0.0, 0.0
	if s.viewLocked() {
		sx, ox, oy = s.fit()
		sy = sx
	} else if s.stretched() {
		w, h := s.frameSize()
		if w > 0 && h > 0 {
			sx *= float64(s.screenW) / float64(w)
			sy *= float64(s.screenH) / float64(h)
		}
	} else {
		// Page.IMEContext has already subtracted the CSS scroll offset.
		ox = float64(s.scrollX) * (sx - 1)
		oy = float64(s.scrollY) * (sy - 1)
	}
	return image.Rect(int(math.Floor(box.X*sx+ox)), int(math.Floor(box.Y*sy+oy)),
		int(math.Ceil((box.X+box.W)*sx+ox)), int(math.Ceil((box.Y+box.H)*sy+oy)))
}
