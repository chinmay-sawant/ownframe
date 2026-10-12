package window

import "github.com/hajimehoshi/ebiten/v2"

// drawBitmapPage draws the cached bitmap over the already filled screen: its
// own pixels for a page larger than the window, scaled to the window
// otherwise. A size that has not been laid out yet is scaled until the new
// picture arrives.
func (s *shell) drawBitmapPage(screen *ebiten.Image, zoom float64) {
	img := s.img
	if img == nil {
		return
	}

	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 || s.screenW == 0 || s.screenH == 0 {
		return
	}

	var op ebiten.DrawImageOptions
	if !s.stretched() {
		if zoom != 1 {
			op.Filter = ebiten.FilterLinear
			op.GeoM.Scale(zoom, zoom)
		}

		op.GeoM.Translate(-float64(s.scrollX), -float64(s.scrollY))
		screen.DrawImage(img, &op)

		return
	}

	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(
		float64(s.screenW)/float64(bounds.Dx())*zoom,
		float64(s.screenH)/float64(bounds.Dy())*zoom,
	)
	screen.DrawImage(img, &op)
}
