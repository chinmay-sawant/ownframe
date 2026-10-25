package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// prewarmMaxPixels bounds the content a first frame may walk to warm the
// raster, text, and chunk caches. A page past it warms the current viewport
// and the next two only, so a huge document does not block its first frame.
const prewarmMaxPixels = 12 << 20

// prewarmViewport paints the scrollable page once into a scratch buffer so the
// raster, chunk, and text caches hold the shapes and runs the first scroll
// will need. The window is already showing its first frame; this runs after
// it, and the scratch image never reaches the screen.
//
// It is the synchronous version of what a compositor's raster worker does.
// Without it, the frame that crosses the viewport overscan pays for every
// first-time shape and text bake in the same frame that rebuilds the buffer,
// which is long enough to read as a scroll freeze.
func (s *shell) prewarmViewport(d *layout.Display, bounds image.Rectangle) {
	if d == nil || bounds.Empty() {
		return
	}
	contentW, contentH := s.contentSize()
	if contentW <= 0 || contentH <= bounds.Dy() {
		return
	}

	lastY := contentH
	if contentW*contentH > prewarmMaxPixels {
		lastY = min(contentH, s.scrollY+2*bounds.Dy())
	}

	scratch := ebiten.NewImage(bounds.Dx(), bounds.Dy())
	defer scratch.Deallocate()

	for y := max(0, s.scrollY-bounds.Dy()); y < lastY; y += bounds.Dy() {
		for x := 0; x < contentW; x += bounds.Dx() {
			scratch.Clear()
			scratch.Fill(s.pageBackground())
			s.renderer.DrawVisibleUnfixedZ(scratch, d, -float64(x), -float64(y), s.viewportPinZ())
		}
	}
}
