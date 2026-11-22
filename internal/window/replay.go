package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

// dropDisplayBuffers releases every retained display-list buffer: the tile
// grid, the viewport cache, and the shared scaled buffer. Frames that paint
// without them (scaled, fitted, bitmap, direct) call it instead of pinning
// a stale copy beside the buffer they actually read.
func (s *shell) dropDisplayBuffers() {
	s.dropTiles()
	s.disposeViewport()
	s.dropScaleBuf()
	s.demoteLayers()
}

// applyRepaint brings the tile grid up to date for this frame. A rebuild
// ensures the grid matches the content and repaints all of it; a rect plan
// repaints only the tiles the rect covers. A blit plan changes no pixels.
func (s *shell) applyRepaint(display *layout.Display, plan repaintPlan) {
	switch plan.mode {
	case repaintBuffer:
		s.ensureTiles(plan.rect.Dx(), plan.rect.Dy())
		s.paintTiles(display, plan.rect)
	case repaintRect:
		s.paintTiles(display, plan.rect)
	}

	s.partial.gen = s.app.Generation()
}

// paintFullReplay paints the whole screen: the page background, then the
// display list. It is the full branch of the replay path, where the tiles
// alone cannot supply the frame.
func (s *shell) paintFullReplay(dst *ebiten.Image, display *layout.Display) {
	dst.Fill(s.pageBackground())
	s.directReplay(dst, display)
}

// frameMarker is a ticking page that can report dirty rects.
type frameMarker interface {
	FrameDirty() bool
}

// captureFrameOptIn records the page's frame-dirty flag for this Update.
// Ebiten can call Draw without an Update in between, and Page.FrameDirty
// clears the flag when it is read, so the window must read it once per tick.
func (s *shell) captureFrameOptIn() {
	if marker, ok := s.app.(frameMarker); ok {
		s.frameOptIn = marker.FrameDirty()
	}
}

// fullReplay reports a ticking page that cannot report dirty rects.
// A Page (frameMarker) stays on the buffer path: Tick marks mutated ops
// itself, and UseFrameDirty is an alias. A no-op tick leaves TakeDirty
// empty so idle park can sleep. A ticker without FrameDirty still replays
// in full.
func (s *shell) fullReplay() bool {
	if !s.isTicking() {
		return false
	}

	_, ok := s.app.(frameMarker)

	return !ok
}

type orderCache struct {
	display  *layout.Display
	safe     bool
	viewport bool
}

// bufferOrderSafe reports that the scroll buffer can precede the viewport
// layers: paint order already puts those layers last.
func (s *shell) bufferOrderSafe() bool {
	s.scanOrder()

	return s.orderCache.safe
}

// hasViewportOps reports a page with fixed or viewport-pinned operations.
// Those layers paint on the screen after the unfixed tile blit, so a
// partial frame keeps the 256 px grid. 1.0 pins chrome with
// SetScrollWindow / SetViewportPinZ.
func (s *shell) hasViewportOps() bool {
	s.scanOrder()

	return s.orderCache.viewport
}

// scanOrder fills the order cache once per display. It records both facts in
// one pass so a partial frame pays for neither separately.
func (s *shell) scanOrder() {
	if s.orderCache.display == s.display {
		return
	}

	s.orderCache = orderCache{display: s.display, safe: true}
	viewportSeen := false

	for _, i := range s.display.Order {
		if i < 0 || i >= len(s.display.Ops) {
			continue
		}

		op := &s.display.Ops[i]
		if op.Kind == layout.DisplayOpNoop || op.Kind == layout.DisplayOpLinkURI {
			continue
		}

		viewport := op.Fixed || s.viewportPinZ() > 0 && op.ZIndexSet && op.ZIndex >= s.viewportPinZ()
		s.orderCache.viewport = s.orderCache.viewport || viewport

		if !viewport && viewportSeen {
			s.orderCache.safe = false
		}

		viewportSeen = viewportSeen || viewport
	}
}

// maxGridViewports bounds the content-sized tile grid against the window.
// The grid holds every content pixel, so it may cost this many viewport
// buffers before the window takes the viewport path, which keeps only a
// window-sized buffer. A gallery of animated cards a few screens tall takes
// the viewport path, where a background blit plus retained layers is cheaper
// than clearing and replaying tiles that hold the same static content.
const maxGridViewports = 3

// oversized reports a content size no Ebiten image can hold. A zero device
// limit means the game is not running, so nothing is oversized. A large but
// holdable size keeps its scale: only the tile grid has a budget.
func oversized(w, h int) bool {
	m := ebiten.MaxImageSize()
	if m <= 0 {
		return false
	}

	return w > m || h > m
}

// gridTooBig reports that the content-sized tile grid would cost far more
// memory than the window keeps. The measure is a ratio of the content area
// to the window area, not a pixel count: a full-screen page on a 4K display
// is large in pixels, but its grid is no bigger than the viewport buffer the
// window keeps either way. Only a page several times the window takes the
// viewport path, which bounds its memory to the window.
func (s *shell) gridTooBig(contentW, contentH int) bool {
	winW, winH := s.screenW, s.screenH
	if contentW <= 0 || contentH <= 0 || winW <= 0 || winH <= 0 {
		return false
	}

	return int64(contentW)*int64(contentH) > int64(maxGridViewports)*int64(winW)*int64(winH)
}

// windowing reports that the page opted into row windowing: it renders only
// the viewport slice between full-height spacers, so the repaint working set
// sizes from the viewport (plus overscan) instead of the full content
// height. Scroll range, spacer layout, and painted output are unchanged;
// only the tile grid, viewport buffers, and warmup sweep shrink. A page
// without the opt-in keeps exactly the old path.
func (s *shell) windowing() bool {
	observer, ok := s.app.(host.ScrollObserver)

	return ok && observer.Windowing()
}

// isTicking reports a screen with a frame callback. A Page marks mutated
// ops itself and stays on the buffer path; a ticker without FrameDirty
// still replays in full.
func (s *shell) isTicking() bool {
	t, ok := s.app.(ticking)

	return ok && t.Ticking()
}

// viewportPinZ is the z layer of the page's viewport-pinned operations.
func (s *shell) viewportPinZ() int {
	if pinner, ok := s.app.(host.ViewportPinner); ok {
		return pinner.ViewportPinZ()
	}
	return 0
}

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
	if rect == dst.Bounds() {
		return dst
	}

	clip, ok := dst.SubImage(rect).(*ebiten.Image)
	if !ok {
		return nil
	}

	return clip
}

// repaintMode is what one frame does with the persistent buffer.
type repaintMode int

const (
	// repaintBuffer repaints the whole buffer.
	repaintBuffer repaintMode = iota
	// repaintRect repaints one dirty rect in the buffer.
	repaintRect
	// repaintBlit leaves the buffer alone. Its pixels are current.
	repaintBlit
)

// repaintPlan is the buffer work for one frame.
type repaintPlan struct {
	mode repaintMode
	rect image.Rectangle
}

// maxRepaintRects caps the dirty rects of one frame. Each rect costs a scan
// of the operation list, so a longer list rebuilds the whole buffer. The page
// tracker keeps up to dirty.MaxRects, and a gallery of animated cards needs
// each one painted on its own or the frame pays for a whole-page replay.
const maxRepaintRects = 24

// planRepaints decides the frame's tile work for content contentW x contentH.
// The content can outgrow the canvas when a fixed root's children overflow, so
// the grid and the full rect use the content. A frame with no damage blits
// when the grid holds this generation and rebuilds it otherwise. A
// full-content rect, a list longer than maxRepaintRects, or a rect that clamps
// away rebuilds the grid. Otherwise every rect gets its own repaint; all
// plans share one mode. clipX is the first page x the devtools dock hides:
// damage right of it needs no pixels, and a clip that grew repaints the strip
// the dock hid.
func planRepaints(contentW, contentH int, st partialState, gen uint64, rects []image.Rectangle, ok bool, clipX int) []repaintPlan {
	full := image.Rect(0, 0, contentW, contentH)
	plans := st.plans[:0]
	needed := len(rects) + 1
	if needed < 1 {
		needed = 1
	}
	if cap(plans) < needed {
		plans = make([]repaintPlan, 0, max(needed, maxRepaintRects+1))
	}

	if full.Empty() || !st.gridCurrent(contentW, contentH) {
		return append(plans[:0], repaintPlan{mode: repaintBuffer, rect: full})
	}

	strip := st.dockStrip(clipX, contentH)
	if !ok || len(rects) == 0 {
		if st.gen != gen {
			return append(plans[:0], repaintPlan{mode: repaintBuffer, rect: full})
		}

		if strip.Empty() {
			return append(plans[:0], repaintPlan{mode: repaintBlit, rect: full})
		}

		return append(plans[:0], repaintPlan{mode: repaintRect, rect: strip})
	}

	for _, rect := range rects {
		inside := rect.Intersect(full)
		if inside.Empty() {
			return append(plans[:0], repaintPlan{mode: repaintBuffer, rect: full})
		}

		visible := clipDockX(inside, clipX)
		if visible.Empty() {
			continue
		}

		if visible == full {
			return append(plans[:0], repaintPlan{mode: repaintBuffer, rect: full})
		}

		plans = append(plans, repaintPlan{mode: repaintRect, rect: visible})
	}

	if !strip.Empty() {
		plans = append(plans, repaintPlan{mode: repaintRect, rect: strip})
	}

	if len(plans) == 0 {
		return append(plans[:0], repaintPlan{mode: repaintBlit, rect: full})
	}

	if len(plans) > maxRepaintRects {
		return append(plans[:0], repaintPlan{mode: repaintBuffer, rect: full})
	}

	return plans
}

// ensureScaleBuf keeps the one shared content-sized buffer both scaled draws
// reuse: the zoomed replay and the fitted replay never run on the same frame,
// so they share one image instead of pinning two content-sized buffers.
func (s *shell) ensureScaleBuf(w, h int) *ebiten.Image {
	if s.scaleBuf == nil || s.scaleBuf.Bounds().Dx() != w || s.scaleBuf.Bounds().Dy() != h {
		if s.scaleBuf != nil {
			s.scaleBuf.Deallocate()
		}

		// This image alternates between a render target and a blit source.
		// Keep it out of the sprite atlas so those switches never migrate it.
		s.scaleBuf = ebiten.NewImageWithOptions(image.Rect(0, 0, w, h), &ebiten.NewImageOptions{Unmanaged: true})
	}

	return s.scaleBuf
}

// dropScaleBuf releases the shared scaled buffer. Tile, viewport, and direct
// frames never read it, so they drop it instead of pinning a stale copy.
func (s *shell) dropScaleBuf() {
	if s.scaleBuf != nil {
		s.scaleBuf.Deallocate()
		s.scaleBuf = nil
	}
}

// drawReplayScaled draws the display list through an offscreen buffer and
// scales it to the window, matching the bitmap path on a resized or clamped
// window.
func (s *shell) drawReplayScaled(dst *ebiten.Image) {
	w, h := s.display.Width, s.display.Height
	viewW, viewH := s.viewSize()
	if w <= 0 || h <= 0 || viewW <= 0 || viewH <= 0 {
		return
	}

	if oversized(w, h) {
		// No image can hold the page, so the draw drops the scale.
		s.dropDisplayBuffers()
		s.directReplay(dst, s.display)

		return
	}

	// The scaled frame reads neither the tile grid nor the viewport cache.
	s.dropDisplayBuffers()
	buf := s.ensureScaleBuf(w, h)

	fillBackground(buf, s.pageBackground())
	s.renderer.DrawVisible(buf, s.display, 0, 0)

	var op ebiten.DrawImageOptions
	op.Filter = ebiten.FilterLinear

	if s.stretched() {
		op.GeoM.Scale(
			float64(viewW)/float64(w)*s.zoom(),
			float64(viewH)/float64(h)*s.zoom(),
		)
	} else {
		op.GeoM.Scale(s.zoom(), s.zoom())
		op.GeoM.Translate(-float64(s.scrollX), -float64(s.scrollY))
	}

	dst.DrawImage(buf, &op)
}

// directReplay draws the list straight to the screen. Ticking pages use it
// because they have no dirty rect; oversized content uses it because no
// buffer can hold it. Viewport-pinned layers draw at the offset the display
// was built with, so they stay at the viewport while the page scrolls.
func (s *shell) directReplay(dst *ebiten.Image, display *layout.Display) {
	s.renderer.DrawVisiblePinned(dst, display,
		-float64(s.scrollX), -float64(s.scrollY),
		s.viewportPinZ(), -float64(s.redrawX), -float64(s.redrawY))
}

// dirtyTaker is the page side of the partial replay. TakeDirtyRects returns
// the regions that changed since the previous call, in CSS pixels, and
// clears them.
type dirtyTaker interface {
	TakeDirtyRects() ([]image.Rectangle, bool)
}

// singleTaker is a page from before the damage list: one dirty rect.
type singleTaker interface {
	TakeDirty() (image.Rectangle, bool)
}

// listTaker adapts a singleTaker to the list method. Its one rect becomes a
// one-element list.
type listTaker struct{ singleTaker }

func (t listTaker) TakeDirtyRects() ([]image.Rectangle, bool) {
	rect, ok := t.TakeDirty()
	if !ok {
		return nil, false
	}

	return []image.Rectangle{rect}, true
}

// damageTaker resolves the page's damage source: the list method when the
// page has it, otherwise a wrapped single rect.
func (s *shell) damageTaker() (dirtyTaker, bool) {
	if taker, ok := s.app.(dirtyTaker); ok {
		return taker, true
	}

	if one, ok := s.app.(singleTaker); ok {
		return listTaker{one}, true
	}

	return nil, false
}
