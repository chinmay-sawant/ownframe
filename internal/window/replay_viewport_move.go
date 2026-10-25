package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// moveViewport keeps the already painted overlap when scrolling crosses the
// overscan. Only the newly exposed strips need a replay of the display list.
func (s *shell) moveViewport(d *layout.Display, previous image.Rectangle) {
	v := &s.viewport
	if v.spare == nil {
		bounds := v.buf.Bounds()
		v.spare = ebiten.NewImageWithOptions(bounds, &ebiten.NewImageOptions{Unmanaged: true})
	}
	v.buf, v.spare = v.spare, v.buf
	v.buf.Clear()
	var options ebiten.DrawImageOptions
	options.GeoM.Translate(float64(previous.Min.X-v.area.Min.X), float64(previous.Min.Y-v.area.Min.Y))
	v.buf.DrawImage(v.spare, &options)
	for _, strip := range viewportExposed(v.area, previous) {
		s.paintViewport(d, strip, s.viewportPinZ())
	}
}

func viewportExposed(next, previous image.Rectangle) [4]image.Rectangle {
	shared := next.Intersect(previous)
	if shared.Empty() {
		return [4]image.Rectangle{next}
	}
	return [4]image.Rectangle{
		image.Rect(next.Min.X, next.Min.Y, next.Max.X, shared.Min.Y),
		image.Rect(next.Min.X, shared.Max.Y, next.Max.X, next.Max.Y),
		image.Rect(next.Min.X, shared.Min.Y, shared.Min.X, shared.Max.Y),
		image.Rect(shared.Max.X, shared.Min.Y, next.Max.X, shared.Max.Y),
	}
}
