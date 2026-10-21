package window

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
)

type bitmapView struct {
	img  *ebiten.Image
	gen  uint64
	area image.Rectangle
	zoom float64
}

func (s *shell) prepareBitmapViewport() error {
	// The provider is window-only; the browser sink serves PNG bytes and
	// never reads retained operations. See host.FallbackDisplayProvider.
	raw, ok := s.app.(host.FallbackDisplayProvider)
	if !s.fallback || !ok || raw.FallbackDisplay() == nil || s.stretched() || s.viewLocked() {
		if s.bitmapView.img != nil {
			s.bitmapView.img.Deallocate()
		}
		s.bitmapView = bitmapView{}
		return nil
	}
	viewW, viewH := s.viewSize()
	area := image.Rect(s.scrollX, s.scrollY, s.scrollX+viewW, s.scrollY+viewH)
	gen, zoom := s.app.Generation(), s.zoom()
	v := &s.bitmapView
	if v.img != nil && v.gen == gen && v.area == area && v.zoom == zoom {
		return nil
	}
	img, err := bitmap.Viewport(raw.FallbackDisplay(), area, zoom)
	if err != nil {
		return err
	}
	if img == nil {
		return nil
	}
	if v.img == nil || v.img.Bounds().Size() != area.Size() {
		if v.img != nil {
			v.img.Deallocate()
		}
		v.img = ebiten.NewImage(area.Dx(), area.Dy())
	}
	pixels := img.(*image.NRGBA)
	v.img.WritePixels(pixels.Pix)
	v.gen, v.area, v.zoom = gen, area, zoom
	return nil
}
