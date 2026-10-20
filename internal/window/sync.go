package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/replay"
)

// syncImage follows the page after each Redraw. A display-list page keeps
// its operations and drops any bitmap; any other page keeps one Ebiten image
// built from its picture. The old image is disposed when it is replaced.
func (s *shell) syncImage() error {
	if s.built() {
		return nil
	}

	front := page.SelectFront(s.app)
	if front.Display != nil {
		s.setDisplay(front.Display)
		s.pullScroll()

		return nil
	}

	if front.Image == nil {
		return errNoImage
	}

	next := ebiten.NewImageFromImage(front.Image)
	s.disposeImage()

	s.img = next
	s.display = nil
	s.fallback = true
	s.seq = front.Generation
	s.pullScroll()

	return nil
}

// built reports that the cached artifact came from the page's current
// generation at the size the page committed, so no rebuild is needed.
func (s *shell) built() bool {
	if s.app.Generation() != s.seq || (s.img == nil && s.display == nil) {
		return false
	}

	w, h := s.app.Size()
	cw, ch := s.canvasSize()

	return cw == w && ch == h
}

func (s *shell) setDisplay(display *layout.Display) {
	s.disposeImage()
	replay.DropStaleChunks(display)

	s.display = display
	s.img = nil
	s.fallback = false
	s.seq = s.app.Generation()
	s.markRedrawOffset()
}

// markRedrawOffset records the scroll offset the current display was built
// with, so viewport-pinned layers can be drawn against it.
func (s *shell) markRedrawOffset() {
	if observer, ok := s.app.(host.ScrollObserver); ok {
		s.redrawX, s.redrawY = observer.ScrollOffset()
	}
}

func (s *shell) disposeImage() {
	if s.img != nil {
		s.img.Dispose()
		s.img = nil
	}
}
