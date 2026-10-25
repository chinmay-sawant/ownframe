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
		// This image alternates between a render target and a scroll source.
		// Keep it out of the sprite atlas so those switches never migrate it.
		s.viewport.buf = ebiten.NewImageWithOptions(
			image.Rect(0, 0, width, height), &ebiten.NewImageOptions{Unmanaged: true},
		)
	}
	top := max(0, s.scrollY-gap)
	s.viewport.area = image.Rect(s.scrollX, top, s.scrollX+width, top+height)
}

// viewportDirtyBounds is the fallback region a viewport repaint covers: the
// bounding box of the list. paintViewportRects uses it only when the list is
// longer than one frame should scan; a short list paints rect by rect.
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

// paintViewportRects repaints each dirty rect in turn. A moving chip then
// costs its own box, not the bounding box of every moving card on screen. A
// list past the frame cap falls back to one bounding-box pass.
func (s *shell) paintViewportRects(d *layout.Display, rects []image.Rectangle) {
	if len(rects) == 0 {
		return
	}
	if len(rects) > maxRepaintRects {
		s.paintViewport(d, viewportDirtyBounds(rects), s.viewportPinZ())

		return
	}
	for _, rect := range rects {
		s.paintViewport(d, rect, s.viewportPinZ())
	}
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
	s.partial.paints++
	s.partial.paintPx += rect.Dx() * rect.Dy()
	clip.Clear()
	clip.Fill(s.pageBackground())
	s.renderer.DrawVisibleUnfixedZBackground(clip, d, -float64(s.viewport.area.Min.X), -float64(s.viewport.area.Min.Y), pinZ)
}
