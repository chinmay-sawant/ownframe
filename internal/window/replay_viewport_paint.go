package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

const viewportOverscan = 256

// viewportSides counts the top and bottom overscan bands.
const viewportSides = 2

func (s *shell) setViewportArea(bounds image.Rectangle) {
	gap := min(viewportOverscan, bounds.Dy())
	width, height := bounds.Dx(), bounds.Dy()+viewportSides*gap
	if limit := ebiten.MaxImageSize(); limit > 0 {
		height = min(height, limit)
	}
	if s.viewport.buf == nil || s.viewport.buf.Bounds().Size() != image.Pt(width, height) {
		s.disposeViewport()
		s.viewport.buf = ebiten.NewImage(width, height)
	}
	top := max(0, s.scrollY-gap)
	s.viewport.area = image.Rect(s.scrollX, top, s.scrollX+width, top+height)
}

// viewportDirtyBounds is the one region a viewport repaint covers: the sole
// dirty rect when the page reports one, otherwise the bounding box of the
// list. The page coalesces nearby rects and caps the list, and paintViewport
// clips the region to the cached area, so the box stays a bounded overdraw.
// The alternative, one clipped replay per rect, would walk the display list
// once per rect instead of once per frame.
func viewportDirtyBounds(rects []image.Rectangle) image.Rectangle {
	if len(rects) == 0 {
		return image.Rectangle{}
	}
	box := rects[0]
	for _, r := range rects[1:] {
		box = box.Union(r)
	}
	return box
}

func (s *shell) paintViewport(d *layout.Display, rect image.Rectangle, pinZ int) {
	rect = rect.Intersect(s.viewport.area).Sub(s.viewport.area.Min)
	if rect.Empty() {
		return
	}
	clip, ok := s.viewport.buf.SubImage(rect).(*ebiten.Image)
	if !ok {
		return
	}
	clip.Clear()
	clip.Fill(s.pageBackground())
	s.renderer.DrawVisibleUnfixedZ(clip, d, -float64(s.viewport.area.Min.X), -float64(s.viewport.area.Min.Y), pinZ)
}
