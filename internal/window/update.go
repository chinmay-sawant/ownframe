package window

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
	"os"
	"time"
)

type updateStage struct {
	name string
	run  func(*shell) error
	// batch shares one redraw batch across the contiguous batched span,
	// so a hover change that coincides with an app refresh lays the page
	// out once instead of twice. The batch opens before the first batched
	// stage runs and flushes after the last one, before syncImage reads
	// the committed generation.
	batch bool
}

func void(f func(*shell)) func(*shell) error {
	return func(s *shell) error { f(s); return nil }
}

var stages = []updateStage{
	{"ctx", func(s *shell) error { return s.ctx.Err() }, false},
	{"a11y", void((*shell).runA11yActions), false},
	{"close", func(s *shell) error {
		if takeCloseRequest(isWSLg() && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)) {
			return ebiten.Termination
		}
		return nil
	}, false},
	{"devSync", func(s *shell) error { return s.devSync() }, false},
	{"ime", func(s *shell) error { return s.imeUpdate() }, false},
	{"keys", func(s *shell) error { return s.keys() }, false},
	{"resize", func(s *shell) error { return s.resize() }, false},
	{"pass", void((*shell).updatePassthrough), false},
	{"pointer", func(s *shell) error { return s.pointer() }, true},
	{"touch", func(s *shell) error { s.stepTouchScroll(time.Now()); return nil }, true},
	{"hold", func(s *shell) error { return s.holdFrame() }, true},
	{"drop", func(s *shell) error { return s.dropPass(ebiten.DroppedFiles()) }, true},
	{"tick", func(s *shell) error { return s.timedErr(&s.dev.stages.tick, s.tickFrame) }, true},
	{"optin", void((*shell).captureFrameOptIn), false},
	{"reload", func(s *shell) error { return s.pollReload() }, false},
	{"sync", func(s *shell) error {
		if err := s.timedErr(&s.dev.stages.sync, s.syncImage); err != nil {
			return err
		}
		return nil
	}, false},
	{"scrollReq", void((*shell).applyScrollRequest), false},
	{"devRefresh", void((*shell).devRefresh), false},
	{"wheel", (*shell).wheel, false},
	{"scrollWin", func(s *shell) error { return s.syncScrollWindow() }, false},
	{"bitmap", func(s *shell) error { return s.prepareBitmapViewport() }, false},
}

func (s *shell) Update() error {
	if s.drawSyncErr != nil {
		err := s.drawSyncErr
		s.drawSyncErr = nil
		return err
	}

	if s.perf {
		start := time.Now()
		defer func() {
			elapsed := time.Since(start)
			s.dev.stages.update = elapsed
			s.dev.stages.pendingUpdate += elapsed
		}()
	}

	if s.app == nil {
		return errNilApp
	}

	// The pointer span resolves hover and gestures and the tick span runs
	// the frame callback; each redraws, so a hover change that coincides
	// with an app refresh lays the page out twice back to back. The
	// batched stages share one redraw batch, and the single flush before
	// syncImage covers both. Batching defers commits, never input: hover,
	// clicks, picks, drops, and the tick callback itself run in order, and
	// the flush lands in the same frame, so hover still paints within one
	// frame. A screen without the batcher role runs every stage directly,
	// exactly as before.
	batcher, ok := s.app.(host.RedrawBatcher)
	batched := false
	// A stage runs app code and can panic. This does not recover: the panic
	// keeps travelling to the crash report. The open batch must not leak
	// into the next frame, so end it without drawing.
	defer func() {
		if batched {
			batched = false
			s.endBatchQuiet(batcher)
		}
	}()
	flush := func() error {
		batched = false
		return batcher.EndRedrawBatch(s.ctx)
	}
	for _, st := range stages {
		if ok && st.batch && !batched {
			batcher.BeginRedrawBatch()
			batched = true
		}
		if batched && !st.batch {
			if err := flush(); err != nil {
				return err
			}
		}
		if err := st.run(s); err != nil {
			// The frame is doomed, so end the batch without drawing:
			// the queued draw stays queued and the next batch commits
			// it against the state that batch changed.
			if batched {
				batched = false
				s.endBatchQuiet(batcher)
			}
			return err
		}
	}
	if batched {
		if err := flush(); err != nil {
			return err
		}
	}
	s.maybePublishDebug()
	return nil
}

// tickFrame runs the screen's per-frame callback when it has one.
func (s *shell) tickFrame() error {
	ticker, ok := s.app.(host.Ticker)
	if !ok {
		return nil
	}
	return ticker.Tick(s.ctx)
}

// endBatchQuiet ends a batch without drawing, so a stage error or panic
// never leaves the deferral open for the next frame. The draw runs with a
// dead context, so the page records nothing, and a queued draw stays queued
// for the next batch. A panic inside the end must not replace the panic
// that is already travelling.
func (s *shell) endBatchQuiet(batcher host.RedrawBatcher) {
	if batcher == nil || s.ctx == nil {
		return
	}
	dead, cancel := context.WithCancel(s.ctx)
	cancel()
	defer func() { _ = recover() }()
	_ = batcher.EndRedrawBatch(dead)
}

// reloadEvery is how often a watched page is asked for file changes.
const reloadEvery = 250 * time.Millisecond

// pollReload asks the screen to reread its files when the interval has
// passed. A change is left to syncImage. An error is printed once and the
// last good frame keeps drawing.
func (s *shell) pollReload() error {
	reloader, ok := s.app.(host.Reloader)
	if !ok {
		return nil
	}

	if !s.lastPoll.IsZero() && time.Since(s.lastPoll) < reloadEvery {
		return nil
	}

	s.lastPoll = time.Now()

	changed, err := reloader.PollReload(s.ctx)
	if err != nil {
		if line, show := s.reloadNote(err); show {
			fmt.Fprintln(os.Stderr, line)
		}

		return nil
	}

	s.lastNote = ""
	if !changed {
		return nil
	}

	s.clampView()

	return s.refreshState()
}

// reloadNote returns the stderr line for err, once per distinct message.
// The text is sanitized so a path cannot forge extra log lines.
func (s *shell) reloadNote(err error) (string, bool) {
	line := "hot reload: " + cleanLog(err.Error())
	if line == s.lastNote {
		return "", false
	}

	s.lastNote = line

	return line, true
}

// clampView pulls the scroll offset back inside a page that got shorter.
func (s *shell) clampView() {
	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	s.scrollX, s.scrollY = clampScroll(s.scrollX, s.scrollY, contentW, contentH, viewW, viewH)
}

// stretched reports that the painted page is being scaled to the viewport.
// The scale is only for the gap while a window or dock drag waits on a
// relayout. Once the page has committed its size, a page that cannot fit the
// viewport (a min-size clamp) is scrolled instead, so a click stays on the
// pixel under the pointer.
func (s *shell) stretched() bool {
	viewW, viewH := s.viewSize()
	frameW, frameH := s.frameSize()
	if frameW == viewW && frameH == viewH {
		return false
	}

	wantW, wantH := s.layoutSize(viewW, viewH)
	haveW, haveH := s.app.Size()
	if haveW == wantW && haveH == wantH {
		return false
	}

	return true
}

func (s *shell) wheel() error {
	if s.devWheel() {
		return nil
	}

	wheelX, wheelY := ebiten.Wheel()
	if wheelX == 0 && wheelY == 0 {
		return nil
	}

	if s.zoomWheel(wheelY) {
		return nil
	}

	if handler, ok := s.app.(host.WheelHandler); ok {
		x, y := s.cursorPosition()
		frameW, frameH := s.frameSize()
		px, py := s.contentAt(x, y, frameW, frameH)

		handled, err := handler.Wheel(s.ctx, px, py, wheelX, wheelY)
		if err != nil {
			return err
		}

		if handled {
			return nil
		}
	}

	if s.stretched() || !s.allowPageScroll() {
		return nil
	}

	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	s.scrollX, s.scrollY = panScroll(
		s.scrollX, s.scrollY, wheelX, wheelY, contentW, contentH, viewW, viewH,
	)

	return nil
}

func (s *shell) frameSize() (int, int) {
	if s.img != nil {
		bounds := s.img.Bounds()

		return bounds.Dx(), bounds.Dy()
	}

	if s.display != nil {
		return s.display.Width, s.display.Height
	}

	return s.app.Size()
}

// drawFitted paints the whole page inside the window.
// One scale serves both axes, and the page stays centered.
func (s *shell) drawFitted(dst *ebiten.Image) {
	scale, ox, oy := s.fit()
	w, h := s.contentSize()
	if w < 1 || h < 1 || scale <= 0 {
		return
	}

	if s.display != nil {
		s.blitFitted(dst, w, h, scale, ox, oy, func(buf *ebiten.Image) {
			s.renderer.Draw(buf, s.display, 0, 0)
		})

		return
	}

	if s.img == nil {
		return
	}

	s.blitFitted(dst, w, h, scale, ox, oy, func(buf *ebiten.Image) {
		buf.DrawImage(s.img, nil)
	})
}

func (s *shell) blitFitted(dst *ebiten.Image, w, h int, scale, ox, oy float64, paint func(*ebiten.Image)) {
	if oversized(w, h) {
		// No image can hold the page, so the fit drops the scale.
		if s.display != nil {
			s.directReplay(dst, s.display)
		}

		return
	}

	// The fitted frame reads neither the tile grid nor the viewport cache.
	s.dropDisplayBuffers()
	buf := s.ensureScaleBuf(w, h)

	buf.Clear()
	buf.Fill(s.pageBackground())
	paint(buf)

	var op ebiten.DrawImageOptions
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(ox, oy)
	dst.DrawImage(buf, &op)
}

func (s *shell) viewLocked() bool {
	lock, ok := s.app.(host.ViewLock)

	return ok && lock.ViewLocked()
}

// holdView drops a pinch. The page stays at its own size.
func (s *shell) holdView() {
	s.fingers.zoom = 1
	s.fingers.span0 = 0
	s.pageZoom = 1
}

func (s *shell) touchZoomAllowed() bool {
	policy, ok := s.app.(host.TouchZoomPolicy)
	return !ok || policy.TouchZoomAllowed()
}

func (s *shell) wheelZoomAllowed() bool {
	policy, ok := s.app.(host.WheelZoomPolicy)
	return !ok || policy.WheelZoomAllowed()
}

func (s *shell) keyZoomAllowed() bool {
	policy, ok := s.app.(host.KeyZoomPolicy)
	return !ok || policy.KeyZoomAllowed()
}

func (s *shell) holdTouchZoom() {
	s.fingers.zoom = 1
	s.fingers.span0 = 0
}

// fitBox scales content into the screen and centers it.
// The scale is the same on both axes, so a tall phone and a desktop window
// both show the whole page.
func fitBox(contentW, contentH, screenW, screenH int) (float64, float64, float64) {
	if contentW < 1 || contentH < 1 || screenW < 1 || screenH < 1 {
		return 1, 0, 0
	}

	scale := math.Min(
		float64(screenW)/float64(contentW),
		float64(screenH)/float64(contentH),
	)
	ox := (float64(screenW) - float64(contentW)*scale) / devCenterDiv
	oy := (float64(screenH) - float64(contentH)*scale) / devCenterDiv

	return scale, ox, oy
}

func (s *shell) fit() (float64, float64, float64) {
	if !s.viewLocked() {
		return 1, 0, 0
	}

	w, h := s.contentSize()
	viewW, viewH := s.viewSize()

	return fitBox(w, h, viewW, viewH)
}

// viewSize is the page viewport: the window minus the devtools dock when the
// overlay is open. The dock shrinks the page like a docked browser inspector
// instead of covering it, so a page relayouts to the strip that stays visible.
func (s *shell) viewSize() (int, int) {
	return s.viewOf(s.screenW, s.screenH)
}

// viewOf is viewSize for an outside size that Layout has not published yet.
func (s *shell) viewOf(w, h int) (int, int) {
	if s.dev.on && s.dev.mobile {
		panel := devMobileRect(w, h, s.dev.mobileExpanded)
		h -= int(panel.H + devMobileSafeBottom)
	} else if s.dev.on {
		dock := int(s.devDockWidth())
		if dock > w {
			dock = w
		}

		w -= dock
	}

	if w < 1 {
		w = 1
	}

	if h < 1 {
		h = 1
	}

	return w, h
}

func (s *shell) layoutSize(w, h int) (int, int) {
	if _, ok := s.app.(host.ViewportSizer); ok && !s.viewLocked() {
		return max(1, w), max(1, h)
	}
	return s.app.Clamp(w, h)
}

func (s *shell) setLayoutSize(w, h int) {
	if viewport, ok := s.app.(host.ViewportSizer); ok {
		viewport.SetViewport(w, h)
		return
	}
	s.app.SetSize(w, h)
}

const scrollStep = 48

// framePoint maps a pointer on the current screen back onto a stretched picture.
// The picture fills the screen, so a point halfway across the window is halfway
// across the picture.
func framePoint(x, y, frameW, frameH, screenW, screenH int) (float64, float64) {
	if frameW <= 0 || frameH <= 0 || screenW <= 0 || screenH <= 0 {
		return 0, 0
	}

	return float64(x) * float64(frameW) / float64(screenW),
		float64(y) * float64(frameH) / float64(screenH)
}

// contentPoint maps a cursor into the painted page.
// When the page is taller than the window, scrollX and scrollY are the
// page position at the window's top left. A stretched picture uses framePoint.
func contentPoint(x, y, scrollX, scrollY int, stretched bool, frameW, frameH, screenW, screenH int) (float64, float64) {
	if stretched {
		return framePoint(x, y, frameW, frameH, screenW, screenH)
	}

	return float64(x + scrollX), float64(y + scrollY)
}

// panScroll moves the window through a page that is larger than the view.
// A positive wheel moves toward the start of that axis.
func panScroll(scrollX, scrollY int, wheelX, wheelY float64, contentW, contentH, viewW, viewH int) (int, int) {
	nextX := int(math.Round(float64(scrollX) - wheelX*scrollStep))
	nextY := int(math.Round(float64(scrollY) - wheelY*scrollStep))

	return clampScroll(nextX, nextY, contentW, contentH, viewW, viewH)
}

func clampScroll(scrollX, scrollY, contentW, contentH, viewW, viewH int) (int, int) {
	return clampAxis(scrollX, contentW-viewW), clampAxis(scrollY, contentH-viewH)
}

func clampAxis(value, limit int) int {
	if limit < 0 {
		limit = 0
	}

	if value < 0 {
		return 0
	}

	if value > limit {
		return limit
	}

	return value
}

// windowBounds returns the largest window the screen asks for, or -1 for no
// bound. A screen without host.MaxSizeProvider is unbounded, as it was
// before the cap became a window bound. The cap is window-only; the browser
// sink serves pictures and never reads it.
func windowBounds(app host.Screen) (int, int) {
	maxW, maxH := -1, -1

	sizer, ok := app.(host.MaxSizeProvider)
	if !ok {
		return maxW, maxH
	}

	maxW, maxH = sizer.MaxSize()
	if maxW <= 0 {
		maxW = -1
	}

	if maxH <= 0 {
		maxH = -1
	}

	return maxW, maxH
}

// boxes returns the active layout boxes without allocating a defensive copy.
// It prefers the current display's box list, then queries host.BoxSlicer if
// supported, and falls back to s.app.Boxes().
func (s *shell) boxes() []layout.Box {
	if s.display != nil && len(s.display.Boxes) > 0 {
		return s.display.Boxes
	}
	if slicer, ok := s.app.(host.BoxSlicer); ok {
		return slicer.BoxesSlice()
	}
	if s.app != nil {
		return s.app.Boxes()
	}

	return nil
}

// boxAt returns the innermost box containing (x, y), or (layout.Box{}, false).
func (s *shell) boxAt(x, y float64) (layout.Box, bool) {
	return devBoxAt(s.boxes(), x, y)
}

// contentSize returns the painted page size, including any box that
// overflows the canvas, so a fixed-width child larger than the window is
// scrollable. The size is cached per page generation: boxes change only on
// Redraw, so repeated calls within one frame cost O(1) on large pages.
// Generation zero never caches, so tests driving a fake screen stay fresh.
func (s *shell) contentSize() (int, int) {
	gen := s.app.Generation()
	if gen != 0 && s.contentGen == gen {
		return s.contentW, s.contentH
	}

	w, h := s.canvasSize()

	viewport := s.viewportOperations()
	for _, b := range s.boxes() {
		if viewportBox(b, viewport) {
			continue
		}

		x2, y2 := b.X+b.W, b.Y+b.H

		if b.HasOverflowClip {
			// Paint clipped by an overflow ancestor never extends the
			// scrollable document: intersect the box with the recorded
			// clip instead of the raw border box. A marquee track twice
			// its lane, for example, must not summon a horizontal
			// scrollbar on every page that clips it.
			c := b.OverflowClip
			x1 := math.Max(b.X, c[0])
			y1 := math.Max(b.Y, c[1])
			x2 = math.Min(x2, c[0]+c[2])
			y2 = math.Min(y2, c[1]+c[3])

			if x2 <= x1 || y2 <= y1 {
				continue
			}
		}

		if right := int(math.Ceil(x2)); right > w {
			w = right
		}

		if bottom := int(math.Ceil(y2)); bottom > h {
			h = bottom
		}
	}

	if gen != 0 {
		s.contentW, s.contentH, s.contentGen = w, h, gen
	}

	return w, h
}

// viewportOperations identifies paint that follows the viewport rather than
// extending the scrollable document.
func (s *shell) viewportOperations() map[uint64]bool {
	if s.display == nil {
		return nil
	}
	pin := s.viewportPinZ()
	ids := make(map[uint64]bool)
	for _, op := range s.display.Ops {
		if op.Fixed || pin > 0 && op.ZIndexSet && op.ZIndex >= pin {
			ids[op.ID] = true
		}
	}
	return ids
}

func viewportBox(b layout.Box, ids map[uint64]bool) bool {
	if len(b.OperationIDs) == 0 {
		return false
	}
	for _, id := range b.OperationIDs {
		if !ids[id] {
			return false
		}
	}
	return true
}

func (s *shell) syncStickyViewport() {
	if !s.viewportSticky {
		return
	}
	observer, ok := s.app.(host.ScrollViewportObserver)
	if !ok {
		return
	}
	width, height := s.viewSize()
	zoom := s.zoom()
	observer.SetScrollViewport(float64(s.scrollX)/zoom, float64(s.scrollY)/zoom, float64(width)/zoom, float64(height)/zoom)
}

// Sticky paint changes with scroll, so its display bypasses cached page tiles.
// Replaying the current viewport clears the previous sticky position as well
// as the new one; ordinary pages retain their existing partial repaint path.
func (s *shell) drawStickyContent(screen *ebiten.Image) {
	s.syncStickyViewport()
	width, height := s.viewSize()
	zoom := s.zoom()
	display := layout.StickyView(s.display, float64(s.scrollX)/zoom, float64(s.scrollY)/zoom, float64(width)/zoom, float64(height)/zoom)
	if display != s.display {
		defer s.renderer.EvictDisplay(display)
	}
	s.partial.live = false
	s.dropDisplayBuffers()
	screen.Fill(s.pageBackground())
	if zoom == 1 && !s.stretched() {
		s.directReplay(screen, display)
		return
	}
	source := s.display
	s.display = display
	s.drawReplayScaled(screen)
	s.display = source
}

// zoom returns the page zoom: the pinch scale times the keyboard or wheel
// zoom, clamped to a usable range. A shell that never zoomed reads 1.
func (s *shell) zoom() float64 {
	if s.viewLocked() {
		return 1
	}

	return clampZoom(s.fingers.zoomOr1() * s.zoomLevel())
}

// zoomLevel returns the keyboard or wheel zoom.
func (s *shell) zoomLevel() float64 {
	if s.pageZoom <= 0 {
		return 1
	}

	return s.pageZoom
}

// zoomBy scales the page zoom by factor.
func (s *shell) zoomBy(factor float64) {
	if s.viewLocked() || factor <= 0 {
		return
	}

	s.pageZoom = clampZoom(s.zoomLevel() * factor)
}

// zoomReset returns the page zoom to 1.
func (s *shell) zoomReset() {
	s.pageZoom = 1
}

// clampZoom keeps a zoom inside the pinch range.
func clampZoom(z float64) float64 {
	if z < minZoom {
		return minZoom
	}

	if z > maxZoom {
		return maxZoom
	}

	return z
}

// contentAt maps a window point into the page for the current scroll and
// pinch zoom.
func (s *shell) contentAt(x, y, frameW, frameH int) (float64, float64) {
	s.syncStickyViewport()
	if s.viewLocked() {
		scale, ox, oy := s.fit()
		return (float64(x) - ox) / scale, (float64(y) - oy) / scale
	}

	viewW, viewH := s.viewSize()

	return contentPointZoom(
		x, y, s.scrollX, s.scrollY, s.zoom(), s.stretched(),
		frameW, frameH, viewW, viewH,
	)
}

// contentPointZoom maps a cursor into the painted page and divides out the
// pinch zoom, so a click stays on the glyph under the finger.
func contentPointZoom(x, y, scrollX, scrollY int, zoom float64, stretched bool, frameW, frameH, screenW, screenH int) (float64, float64) {
	px, py := contentPoint(x, y, scrollX, scrollY, stretched, frameW, frameH, screenW, screenH)
	if zoom <= 0 {
		zoom = 1
	}

	return px / zoom, py / zoom
}

// zoomStep is one keyboard zoom press or one wheel notch.
const zoomStep = 1.1

// zoomModifier reports that Control or Command is down, and not AltGr.
func zoomModifier() bool {
	command := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	altGr := ebiten.IsKeyPressed(ebiten.KeyAlt) && !ebiten.IsKeyPressed(ebiten.KeyMeta)

	return command && !altGr
}

// zoomKey applies one Control+=, Control+-, or Control+0 press. It reports
// whether the window consumed the key; the plain keys stay with the page.
// A page that refuses key zoom keeps the chords, so the page can run its
// own layout zoom (the code view does); the painted scale stays untouched.
func (s *shell) zoomKey(key ebiten.Key, down bool, mods modifiers) bool {
	if s.viewLocked() || !mods.command() || mods.Alt {
		return false
	}
	if !s.keyZoomAllowed() {
		return false
	}

	switch key {
	case ebiten.KeyEqual, ebiten.KeyKPAdd:
		if down {
			s.zoomBy(zoomStep)
		}
	case ebiten.KeyMinus, ebiten.KeyKPSubtract:
		if down {
			s.zoomBy(1 / zoomStep)
		}
	case ebiten.Key0, ebiten.KeyKP0:
		if down {
			s.zoomReset()
		}
	default:
		return false
	}

	return true
}

// wheelZoomFactor returns the factor for one wheel delta. Half a notch on a
// trackpad moves the zoom half as far.
func wheelZoomFactor(wheelY float64) float64 {
	if wheelY == 0 {
		return 1
	}

	return math.Pow(zoomStep, wheelY)
}

// zoomWheel applies Control+wheel to the page zoom and reports whether the
// wheel went to the zoom instead of the scroll.
// wheelMayZoom reports that this notch is allowed to scale the page. The
// modifier check stays in zoomWheel because it reads the live key state.
func (s *shell) wheelMayZoom(wheelY float64) bool {
	return !s.viewLocked() && wheelY != 0 && s.wheelZoomAllowed()
}

func (s *shell) zoomWheel(wheelY float64) bool {
	if !s.wheelMayZoom(wheelY) || !zoomModifier() {
		return false
	}

	s.zoomBy(wheelZoomFactor(wheelY))

	return true
}
