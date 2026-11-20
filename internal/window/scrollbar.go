package window

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"runtime"
)

const (
	scrollbarThickness = 8
	scrollbarMinThumb  = 24
)

const scrollThumbHex = 0xbdbdbdff

var scrollThumb = devRGBA(scrollThumbHex)

func (s *shell) canvasSize() (int, int) {
	if s.display != nil {
		return s.display.Width, s.display.Height
	}

	if s.img != nil {
		b := s.img.Bounds()

		return b.Dx(), b.Dy()
	}

	return s.app.Size()
}

// scrollbarVisible reports whether content overflows the viewport.
func scrollbarVisible(content, viewport int) bool {
	return content > viewport && viewport > 0
}

// drawScrollbars overlays the thumbs a page larger than the window needs.
// Only the thumb is painted, so the page background stays visible.
func (s *shell) drawScrollbars(dst *ebiten.Image) {
	if !s.scrollbarsEnabled() {
		return
	}

	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	thick := float32(scrollbarThickness)

	if scrollbarVisible(contentH, viewH) {
		pos, length := scrollbarThumb(viewH, contentH, viewH, s.scrollY)
		x := float32(viewW) - thick
		vector.FillRect(dst, x, pos, thick, length, scrollThumb, false)
		s.frameNote(frameRect(float64(x), float64(pos), float64(thick), float64(length)))
	}

	if scrollbarVisible(contentW, viewW) {
		pos, length := scrollbarThumb(viewW, contentW, viewW, s.scrollX)
		y := float32(viewH) - thick
		vector.FillRect(dst, pos, y, length, thick, scrollThumb, false)
		s.frameNote(frameRect(float64(pos), float64(y), float64(length), float64(thick)))
	}
}

func (s *shell) dragScrollbar(x, y int) {
	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()

	switch s.dragAxis {
	case dragVertical:
		pos := float64(y) - s.dragGrab
		s.scrollY = clampAxis(scrollbarOffset(pos, viewH, contentH, viewH), contentH-viewH)
	case dragHorizontal:
		pos := float64(x) - s.dragGrab
		s.scrollX = clampAxis(scrollbarOffset(pos, viewW, contentW, viewW), contentW-viewW)
	}
}

const (
	dragNone = iota
	dragVertical
	dragHorizontal
)

// pointerScrollbar handles presses and drags on a scrollbar thumb. It reports
// whether the event was consumed, so content clicks are skipped.
func (s *shell) pointerScrollbar(x, y int, down, pressed bool) bool {
	if !s.scrollbarsEnabled() {
		s.dragAxis = dragNone
		return false
	}

	if !down || s.dragAxis == dragNone {
		if axis, grab, ok := s.scrollbarGrab(x, y); ok {
			s.applyThumbCursor(axis)

			if pressed {
				s.dragAxis = axis
				s.dragGrab = grab

				return true
			}
		}
	}

	if !down {
		s.dragAxis = dragNone

		return false
	}

	if s.dragAxis == dragNone {
		return false
	}

	s.dragScrollbar(x, y)

	return true
}

// scrollbarGrab reports the drag axis and the cursor offset inside the thumb
// when the point sits on a scrollbar thumb.
func (s *shell) scrollbarGrab(x, y int) (int, float64, bool) {
	if !s.scrollbarsEnabled() {
		return dragNone, 0, false
	}

	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()

	if scrollbarVisible(contentH, viewH) && x >= viewW-scrollbarThickness {
		pos, length := scrollbarThumb(viewH, contentH, viewH, s.scrollY)
		if float64(y) >= float64(pos) && float64(y) <= float64(pos+length) {
			return dragVertical, float64(y) - float64(pos), true
		}
	}

	if scrollbarVisible(contentW, viewW) && y >= viewH-scrollbarThickness {
		pos, length := scrollbarThumb(viewW, contentW, viewW, s.scrollX)
		if float64(x) >= float64(pos) && float64(x) <= float64(pos+length) {
			return dragHorizontal, float64(x) - float64(pos), true
		}
	}

	return dragNone, 0, false
}

func (s *shell) scrollbarsEnabled() bool {
	if s.stretched() || !s.allowPageScroll() {
		return false
	}

	mode := host.ScrollbarAuto
	if provider, ok := s.app.(host.ScrollbarModeProvider); ok {
		mode = provider.ScrollbarMode()
	}

	return scrollbarModeEnabled(mode, runtime.GOOS)
}

func scrollbarModeEnabled(mode host.ScrollbarMode, goos string) bool {
	switch mode {
	case host.ScrollbarVisible:
		return true
	case host.ScrollbarHidden:
		return false
	default:
		return goos != "android" && goos != "ios"
	}
}

// scrollbarThumb returns the thumb position and length along a track.
func scrollbarThumb(track, content, viewport, offset int) (float32, float32) {
	length := float32(track) * float32(viewport) / float32(content)
	if length < scrollbarMinThumb {
		length = scrollbarMinThumb
	}

	if length > float32(track) {
		length = float32(track)
	}

	maxOffset := content - viewport
	pos := float32(0)

	if maxOffset > 0 && float32(track) > length {
		pos = (float32(track) - length) * float32(offset) / float32(maxOffset)
	}

	return pos, length
}

// scrollbarOffset maps a thumb position back to a scroll offset.
func scrollbarOffset(pos float64, track, content, viewport int) int {
	_, length := scrollbarThumb(track, content, viewport, 0)
	span := float64(track) - float64(length)
	maxOffset := content - viewport

	if maxOffset <= 0 || span <= 0 {
		return 0
	}

	if pos < 0 {
		pos = 0
	}

	if pos > span {
		pos = span
	}

	return int(pos * float64(maxOffset) / span)
}

// pullScroll clamps the scroll offset to the content the latest relayout
// produced, so a page that got shorter cannot show empty space below it.
func (s *shell) pullScroll() {
	if !s.allowPageScroll() {
		s.scrollX, s.scrollY = 0, 0
		return
	}
	if s.stretched() {
		return
	}

	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	s.scrollX, s.scrollY = clampScroll(s.scrollX, s.scrollY, contentW, contentH, viewW, viewH)
}

func (s *shell) allowPageScroll() bool {
	if s.viewLocked() {
		return false
	}
	gate, ok := s.app.(host.ScrollGate)

	if !ok {
		return true
	}

	return gate.AllowScroll()
}

// applyScrollRequest consumes a page scroll request and moves the offset.
// The shell keeps ownership of the offset; the request is clamped to the
// content and cleared either way.
func (s *shell) applyScrollRequest() {
	requester, ok := s.app.(host.ScrollRequester)
	if !ok {
		return
	}

	req, ok := requester.TakeScroll()
	if !ok {
		return
	}
	if !s.allowPageScroll() {
		s.scrollX, s.scrollY = 0, 0
		return
	}

	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	if req.Absolute {
		s.scrollX, s.scrollY = clampScroll(req.X, req.Y, contentW, contentH, viewW, viewH)

		return
	}

	s.scrollX, s.scrollY = clampScroll(
		s.scrollX+req.X, s.scrollY+req.Y, contentW, contentH, viewW, viewH,
	)
}

// syncScrollWindow reports offset changes to a virtualized page. The shell
// keeps owning the offset; the page slices its rows and redraws only when
// it opted into windowing. Every other page scrolls by blitting, as before.
func (s *shell) syncScrollWindow() error {
	s.syncStickyViewport()
	observer, ok := s.app.(host.ScrollObserver)
	if !ok {
		return nil
	}

	x, y := observer.ScrollOffset()
	if x == s.scrollX && y == s.scrollY {
		return nil
	}

	observer.SetScrollOffset(s.scrollX, s.scrollY)
	if !observer.Windowing() {
		return nil
	}

	if stepper, ok := s.app.(host.ScrollWindowStepper); ok && !stepper.StepScrollWindow() {
		return nil
	}

	if err := s.app.Redraw(s.ctx); err != nil {
		return err
	}
	if s.app.Display() != nil {
		return s.syncImage()
	}
	return nil
}
