package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// paintPinnedChrome draws fixed and viewport-pinned operations on the
// screen after the unfixed tile blit. Tiles never hold those layers, so
// a scroll or a dirty message rect would otherwise leave stale chrome.
func (s *shell) paintPinnedChrome(dst *ebiten.Image, d *layout.Display) {
	if dst == nil || d == nil || !s.hasViewportOps() {
		return
	}

	clip := s.pinClip(dst)
	if clip == nil {
		return
	}

	s.partial.pins++
	s.renderer.DrawFixed(clip, d)
	s.renderer.DrawPinnedZ(clip, d, s.viewportPinZ(),
		-float64(s.redrawX), -float64(s.redrawY))
}

// pinClip is the destination strip the dock leaves visible. Pinned
// chrome uses screen space; viewSize already subtracts the dock.
func (s *shell) pinClip(dst *ebiten.Image) *ebiten.Image {
	viewW, viewH := s.viewSize()
	rect := image.Rect(0, 0, viewW, viewH).Intersect(dst.Bounds())
	if rect.Empty() {
		return nil
	}

	clip, ok := dst.SubImage(rect).(*ebiten.Image)
	if !ok {
		return nil
	}

	return clip
}
