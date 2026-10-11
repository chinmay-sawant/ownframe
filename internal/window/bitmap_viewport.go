package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/hajimehoshi/ebiten/v2"
)

type fallbackDisplay interface{ FallbackDisplay() *layout.Display }
type bitmapView struct {
	img  *ebiten.Image
	gen  uint64
	area image.Rectangle
	zoom float64
}

func (s *shell) prepareBitmapViewport() error {
	raw, ok := s.app.(fallbackDisplay)
	if !s.fallback || !ok || raw.FallbackDisplay() == nil || s.stretched() || s.viewLocked() {
		if s.bitmapView.img != nil {
			s.bitmapView.img.Dispose()
		}
		s.bitmapView = bitmapView{}
		return nil
	}
	area := image.Rect(s.scrollX, s.scrollY, s.scrollX+s.screenW, s.scrollY+s.screenH)
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
			v.img.Dispose()
		}
		v.img = ebiten.NewImage(area.Dx(), area.Dy())
	}
	pixels := img.(*image.NRGBA)
	v.img.WritePixels(pixels.Pix)
	v.gen, v.area, v.zoom = gen, area, zoom
	return nil
}
