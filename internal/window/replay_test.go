package window

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"slices"
	"testing"
)

// TestOversizedIgnoresLargePage checks that only a size no image can hold
// counts as oversized. A large page keeps its scale on the scaled and the
// fitted paths.
func TestOversizedIgnoresLargePage(t *testing.T) {
	t.Parallel()

	if oversized(1920, 2200) {
		t.Fatal("a large page must not count as oversized")
	}
	if oversized(3840, 2160) {
		t.Fatal("a 4K page must not count as oversized")
	}
}

// TestGridTooBigMeasuresTheWindow checks the tile grid budget. It is a
// ratio of the content area to the window area, so a page a few screens
// tall keeps the grid and a page many times the window does not.
func TestGridTooBigMeasuresTheWindow(t *testing.T) {
	t.Parallel()

	s := &shell{screenW: 1920, screenH: 1080}
	if s.gridTooBig(1920, 2200) {
		t.Fatal("a page twice the window must keep the tile grid")
	}
	if !s.gridTooBig(1920, 19200) {
		t.Fatal("a page ten times the window must take the viewport path")
	}

	k := &shell{screenW: 3840, screenH: 2160}
	if k.gridTooBig(3840, 2160) {
		t.Fatal("a full-screen 4K page must keep the tile grid")
	}
	if !k.gridTooBig(3840, 43200) {
		t.Fatal("a 4K page ten times the window must take the viewport path")
	}
}

// TestGridTooBigNeedsAWindow checks that a shell without a window size has
// no budget, so the content grid is measured against nothing.
func TestGridTooBigNeedsAWindow(t *testing.T) {
	t.Parallel()

	s := &shell{}
	if s.gridTooBig(4000, 4000) {
		t.Fatal("a shell without a window size must have no budget")
	}
}

// TestPartialBudgetKeepsTiledPath checks that an ordinary large page keeps
// the tile grid and reports an idle frame. The page is over the old four
// megabyte budget, so an absolute count would have moved it to the viewport
// path.
func TestPartialBudgetKeepsTiledPath(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 2200)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(1920, 1080)

	s.drawReplayPartial(dst)
	s.drawReplayPartial(dst)

	if s.viewport.buf != nil {
		t.Fatal("a large page took the viewport path")
	}
	if s.partial.tiles == nil {
		t.Fatal("no content grid")
	}
	if !s.idleCleanFrame() {
		t.Fatal("a large page never reports an idle frame")
	}
}

// TestPartialBudgetTakesViewportPath checks that a page far bigger than the
// window takes the viewport path, which keeps its memory bounded.
func TestPartialBudgetTakesViewportPath(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 22000)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(1920, 1080)

	s.drawReplayPartial(dst)

	if s.viewport.buf == nil {
		t.Fatal("a huge page kept the content grid")
	}
	if s.partial.tiles != nil {
		t.Fatal("the content grid was retained")
	}
}

// TestPartialBudgetViewportIdles checks that a stationary viewport blit can
// park. A scroll inside the overscan, a generation change, a damaging tick,
// or a dirty rect keeps the frame clock; a quiet tick idles.
func TestPartialBudgetViewportIdles(t *testing.T) {
	t.Parallel()

	// Desktop Run turns clearing off. A host that clears every frame marks
	// the blit stale, so this test matches that desktop setting.
	ebiten.SetScreenClearedEveryFrame(false)

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 22000)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(1920, 1080)

	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.viewport.buf == nil {
		t.Fatal("a huge page kept the content grid")
	}
	if s.idleCleanFrame() {
		t.Fatal("the first viewport frame must read dirty")
	}

	screen.ok = false
	settleViewportWarmup(s, dst)
	if !s.idleCleanFrame() {
		t.Fatal("a stable viewport blit must read clean")
	}

	s.scrollY = 8
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a scroll inside the overscan must read dirty")
	}

	s.drawReplayPartial(dst)
	s.frameCommit()
	if !s.idleCleanFrame() {
		t.Fatal("the same scroll must read clean")
	}

	screen.gen++
	s.seq = screen.gen
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a generation change must read dirty")
	}

	settleViewportWarmup(s, dst)
	if !s.idleCleanFrame() {
		t.Fatal("the new generation with no dirty must read clean")
	}

	screen.ticking = true
	s.frameOptIn = false
	s.drawReplayPartial(dst)
	s.frameCommit()
	if !s.idleCleanFrame() {
		t.Fatal("a quiet ticking viewport must read clean")
	}

	s.frameOptIn = true
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a ticking viewport that set FrameDirty must read dirty")
	}

	s.frameOptIn = false
	screen.ticking = false
	s.drawReplayPartial(dst)
	s.frameCommit()
	if !s.idleCleanFrame() {
		t.Fatal("a quiet viewport must read clean after ticking stops")
	}

	screen.ok = true
	screen.rects = []image.Rectangle{image.Rect(10, 10, 30, 30)}
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a dirty rect must read dirty")
	}
}

// Warmup progresses over frames before a static page can park.
func settleViewportWarmup(s *shell, dst *ebiten.Image) {
	for range 100 {
		s.drawReplayPartial(dst)
		s.frameCommit()
		if s.idleCleanFrame() {
			return
		}
	}
}

// lockedScreen is a page that fills the window, so the fitted path runs.
type lockedScreen struct {
	*dirtyScreen
}

func (l *lockedScreen) ViewLocked() bool { return true }

// TestScaledPathKeepsLargePage checks that a large page draws through the
// scaled buffer, so the zoom keeps its scale.
func TestScaledPathKeepsLargePage(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 2200)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(64, 64)

	s.pageZoom = 1.25
	s.drawContent(dst)

	if s.scaleBuf == nil {
		t.Fatal("a large page fell back to a direct replay")
	}
	if got := s.scaleBuf.Bounds().Size(); got != image.Pt(1920, 2200) {
		t.Fatalf("buffer = %v", got)
	}
	s.scaleBuf.Deallocate()
}

// TestFittedPathKeepsLargePage checks that a large page draws through the
// fitted buffer, so the fit keeps its scale.
func TestFittedPathKeepsLargePage(t *testing.T) {
	t.Parallel()

	inner := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	inner.display = partialDisplay(1920, 2200)
	inner.gen = 1
	app := &lockedScreen{dirtyScreen: inner}
	s := &shell{
		app: app, ctx: context.Background(), display: inner.display,
		seq: inner.gen, screenW: 1920, screenH: 1080,
	}
	dst := partialBuf(64, 64)

	s.drawContent(dst)

	if s.scaleBuf == nil {
		t.Fatal("a large page fell back to a direct replay")
	}
	if got := s.scaleBuf.Bounds().Size(); got != image.Pt(1920, 2200) {
		t.Fatalf("buffer = %v", got)
	}
	s.scaleBuf.Deallocate()
}

// TestBufferCoversContentPastTheCanvas pins the scrolling example: the
// laid-out canvas is the window size, but the page's boxes reach past it.
// The tile grid and the full plan must cover the content, or scrolling
// paints blank.
func TestBufferCoversContentPastTheCanvas(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.boxes = []layout.Box{{X: 0, Y: 0, W: 100, H: 300}}
	screen.gen = 1
	s := newDirtyShell(screen)

	s.drawReplayPartial(partialBuf(100, 100))

	if s.partial.tiles == nil {
		t.Fatal("no tile grid for the taller content")
	}

	if got := s.partial.tiles.height; got != 300 {
		t.Fatalf("grid height = %d, want the 300 px content", got)
	}

	if len(s.partial.last) != 1 || s.partial.last[0].Dy() != 300 {
		t.Fatalf("full plan rects = %v, want one 300 px rect", s.partial.last)
	}
}

// Two separated changes in one frame repaint as two rects, not their union.
func TestTwoRectsRepaintSeparately(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	first, second := image.Rect(10, 10, 30, 30), image.Rect(60, 60, 80, 80)
	screen.layouts++
	screen.rects, screen.ok = []image.Rectangle{first, second}, true
	screen.gen++
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 2 ||
		s.partial.last[0] != first || s.partial.last[1] != second {
		t.Fatalf("mode = %v, rects = %v, want %v plus %v",
			s.partial.mode, s.partial.last, first, second)
	}
}

// A page with only TakeDirty still lands on the per-rect path: the window
// wraps its one rect as a one-element list.
func TestSingleRectTakerStillRepaints(t *testing.T) {
	t.Parallel()

	screen := &oldScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := &shell{
		app: screen, ctx: context.Background(),
		display: screen.display, seq: screen.gen,
		screenW: screen.width, screenH: screen.height,
	}
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	if s.partial.tiles == nil {
		t.Fatal("no tile grid for a single-rect page")
	}

	rect := image.Rect(10, 10, 30, 30)
	screen.gen++
	screen.rect, screen.ok = rect, true
	s.drawReplayPartial(dst)

	if screen.takes != 2 {
		t.Fatalf("TakeDirty calls = %d, want 2", screen.takes)
	}

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != rect {
		t.Fatalf("mode = %v, rects = %v, want %v", s.partial.mode, s.partial.last, rect)
	}
}

// dockShell is a page with the devtools dock open. A 600 px window gives the
// panel the 220 px minimum, so the page clips at x 380.
func dockShell(width, height int) (*dirtyScreen, *shell) {
	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: width, height: height}}
	screen.display = partialDisplay(width, height)
	screen.gen = 1
	s := newDirtyShell(screen)
	s.dev.on = true

	return screen, s
}

// TestDockClipDropsHiddenDamage checks a dirty rect fully under the dock
// repaints no tile and blits none.
func TestDockClipDropsHiddenDamage(t *testing.T) {
	t.Parallel()

	screen, s := dockShell(600, 300)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.partial.paints != 4 || s.partial.paintPx != 380*300 {
		t.Fatalf("rebuild = %d tiles %d px, want the 4 left of the dock",
			s.partial.paints, s.partial.paintPx)
	}

	screen.click(image.Rect(400, 10, 420, 20))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintBlit || s.partial.paints != 0 || s.partial.blits != 0 {
		t.Fatalf("hidden damage: mode = %v, paints = %d, blits = %d, want a bare blit",
			s.partial.mode, s.partial.paints, s.partial.blits)
	}
}

// TestDockClipTrimsPartialTiles checks a tile the dock half covers repaints
// and blits only its visible part.
func TestDockClipTrimsPartialTiles(t *testing.T) {
	t.Parallel()

	screen, s := dockShell(600, 300)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	s.frameCommit()

	screen.click(image.Rect(250, 10, 420, 20))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != image.Rect(250, 10, 380, 20) {
		t.Fatalf("plan = %v %v, want the visible part", s.partial.mode, s.partial.last)
	}

	if s.partial.paintPx != 1300 || s.partial.blits != 2 || s.partial.blitPx != 1300 {
		t.Fatalf("paint = %d px, blits = %d (%d px), want 1300 px in two tile blits",
			s.partial.paintPx, s.partial.blits, s.partial.blitPx)
	}
}

// TestDockCloseRepaintsHiddenStrip checks closing the dock repaints the page
// strip the panel hid, even though the page reports no damage.
func TestDockCloseRepaintsHiddenStrip(t *testing.T) {
	t.Parallel()

	screen, s := dockShell(600, 300)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	s.frameCommit()

	screen.click(image.Rect(300, 10, 320, 20))
	s.drawReplayPartial(dst)

	s.dev.on = false
	screen.rects, screen.ok = nil, false
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != image.Rect(380, 0, 600, 300) {
		t.Fatalf("close = %v %v, want the strip the dock hid", s.partial.mode, s.partial.last)
	}

	if s.partial.paintPx != 220*300 || s.partial.paints != 4 {
		t.Fatalf("close painted %d px in %d tiles, want the 220 px strip",
			s.partial.paintPx, s.partial.paints)
	}
}

// TestDockClipIgnoredWhenClosed checks a closed dock clips nothing: damage
// where the panel would sit repaints in full.
func TestDockClipIgnoredWhenClosed(t *testing.T) {
	t.Parallel()

	screen, s := dockShell(600, 300)
	s.dev.on = false
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.partial.paints != 6 || s.partial.paintPx != 600*300 {
		t.Fatalf("rebuild = %d tiles %d px, want the whole content",
			s.partial.paints, s.partial.paintPx)
	}

	screen.click(image.Rect(300, 10, 320, 20))
	s.drawReplayPartial(dst)

	if s.partial.paintPx != 20*10 || s.partial.blits != 1 {
		t.Fatalf("closed damage = %d px in %d blits, want 200 px in one",
			s.partial.paintPx, s.partial.blits)
	}
}

// dockDamageFrame runs a rebuild and then one damage frame on a page of w x h
// with the dock open or closed. It returns the damage frame's painted area
// and blit count plus the rebuild's painted tiles and area.
func dockDamageFrame(w, h int, dock bool, damage image.Rectangle) (px, blits, tiles, tilePx int) {
	screen, s := dockShell(w, h)
	s.dev.on = dock
	dst := partialBuf(w, h)

	s.drawReplayPartial(dst)
	s.frameCommit()
	tiles, tilePx = s.partial.paints, s.partial.paintPx

	screen.click(damage)
	s.drawReplayPartial(dst)

	return s.partial.paintPx, s.partial.blits, tiles, tilePx
}

// TestDockClipMeasuresTheWork pins the dock clip's savings on a board-sized
// page: a rebuild and a fixed damage rect, closed and open. The numbers go to
// the log for the perf report.
func TestDockClipMeasuresTheWork(t *testing.T) {
	t.Parallel()

	const w, h = 1908, 1080
	damage := image.Rect(1400, 200, 1908, 500)

	cPx, cBlits, cTiles, cTilePx := dockDamageFrame(w, h, false, damage)
	oPx, oBlits, oTiles, oTilePx := dockDamageFrame(w, h, true, damage)

	t.Logf("dock closed: rebuild %d tiles / %d px; damage %d px in %d blits",
		cTiles, cTilePx, cPx, cBlits)
	t.Logf("dock open:   rebuild %d tiles / %d px; damage %d px in %d blits",
		oTiles, oTilePx, oPx, oBlits)

	if cTiles != 40 || cTilePx != w*h || cPx != 508*300 || cBlits != 6 {
		t.Fatalf("closed = rebuild %d tiles / %d px, damage %d px in %d blits",
			cTiles, cTilePx, cPx, cBlits)
	}

	if oTiles != 35 || oTilePx != 1568*h || oPx != 168*300 || oBlits != 4 {
		t.Fatalf("open = rebuild %d tiles / %d px, damage %d px in %d blits",
			oTiles, oTilePx, oPx, oBlits)
	}
}

func TestPartialPathNeedsATaker(t *testing.T) {
	t.Parallel()

	plain := &fakeScreen{width: 100, height: 100}
	s := &shell{
		app: plain, display: partialDisplay(100, 100), seq: 1,
		screenW: 100, screenH: 100,
	}

	s.drawReplayPartial(partialBuf(100, 100))

	if s.partial.tiles != nil {
		t.Fatal("grid built without a dirty taker")
	}
}

func TestTickingPageKeepsBufferPath(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.ticking = true
	s := newDirtyShell(screen)

	s.drawReplayPartial(partialBuf(100, 100))

	if screen.takes == 0 {
		t.Fatal("ticking Page must take damage rects")
	}

	if s.partial.tiles == nil {
		t.Fatal("grid missing for a ticking Page")
	}
}

// TestScrollOnlyFrameBlitsWholeBuffer checks that a committed frame leaves a
// scroll-only frame one whole-buffer copy at the new offset: the repaintBlit
// plan contributes no rects, so the whole flag supplies the copy.
func TestScrollOnlyFrameBlitsWholeBuffer(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.Draw(dst)
	s.scrollY = 20
	screen.ok = false
	s.Draw(dst)

	if s.partial.mode != repaintBlit {
		t.Fatalf("scroll mode = %v, want blit", s.partial.mode)
	}

	if !s.partial.whole {
		t.Fatal("scroll frame did not set whole")
	}

	if full, rects := blitPlan(s.partial.plans, s.partial.chrome); full || len(rects) != 0 {
		t.Fatalf("blit plan = %v (full = %v), want no rects; whole supplies the copy", rects, full)
	}
}

// TestFrameDirtyKeepsTheBuffer checks a ticking page that opted into dirty
// rects builds the buffer once and then leaves the retained screen alone: an
// idle frame plans a blit that copies nothing.
func TestFrameDirtyKeepsTheBuffer(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	screen.ticking = true
	screen.frame = true
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	// The Update captures the opt-in. The draws below model Ebiten issuing a
	// second Draw without another Update: the captured value must stand.
	s.captureFrameOptIn()

	s.Draw(dst)

	if s.partial.tiles == nil {
		t.Fatal("opted-in tick did not build the tile grid")
	}

	s.Draw(dst)

	if s.partial.mode != repaintBlit {
		t.Fatalf("second frame mode = %v, want blit", s.partial.mode)
	}

	if s.partial.whole {
		t.Fatal("idle frame set whole without a scroll")
	}

	full, rects := blitPlan(s.partial.plans, s.partial.chrome)
	if full || len(rects) != 0 {
		t.Fatalf("idle frame blits = %v (full = %v), want none", rects, full)
	}
}

// layerDirtyScreen is a dirtyScreen that also names one promoted element.
type layerDirtyScreen struct {
	*dirtyScreen
	specs []replay.LayerSpec
}

func (s *layerDirtyScreen) LayerSpecs(*layout.Display) []replay.LayerSpec {
	return s.specs
}

// windowLayerDisplay is a 100x100 page with a two-op card at (10,10).
func windowLayerDisplay() *layout.Display {
	const pt = 72.0 / 96.0

	return &layout.Display{
		Width: 100, Height: 100, PointsPerPixel: pt, PixelPerPoint: 1,
		Boxes: []layout.Box{
			{ID: "page", X: 0, Y: 0, W: 100, H: 100},
			{ID: "card", X: 10, Y: 10, W: 20, H: 20, ParentIndex: 0},
		},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 100 * pt, H: 100 * pt, R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10 * pt, Y: 10 * pt, W: 20 * pt, H: 10 * pt, R: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10 * pt, Y: 20 * pt, W: 20 * pt, H: 10 * pt, B: 1, Alpha: 1},
		},
		Order:    []int{0, 1, 2},
		OpOwners: []int{-1, 1, 1},
	}
}

func windowLayerMatrix(tx float64) layout.Matrix2D {
	return layout.Matrix2D{A: 1, D: 1, E: tx}
}

// TestPartialFrameAddsLayerDamage pins the window seam: a clean page frame
// with a moving retained layer must still repaint the layer's old ∪ new pose,
// and dropping the layer must repaint the pose it leaves behind.
func TestPartialFrameAddsLayerDamage(t *testing.T) {
	d := windowLayerDisplay()
	app := &layerDirtyScreen{
		dirtyScreen: &dirtyScreen{
			fakeScreen: &fakeScreen{width: 100, height: 100},
			display:    d,
			gen:        1,
			ok:         true,
		},
	}
	shell := newDirtyShell(app.dirtyScreen)
	shell.app = app
	shell.display = d
	dst := ebiten.NewImage(100, 100)
	defer dst.Deallocate()

	// The first frame builds the tile grid. Promotion then adds its own
	// damage to an otherwise clean page frame.
	shell.drawReplayPartial(dst)
	app.specs = []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: windowLayerMatrix(2), Alpha: 1}}
	shell.drawReplayPartial(dst)
	if !shell.renderer.LayersActive(d) {
		t.Fatal("the window did not activate the promoted layer")
	}
	if shell.partial.mode != repaintRect || len(shell.partial.last) == 0 {
		t.Fatalf("promotion mode = %v, last = %v", shell.partial.mode, shell.partial.last)
	}

	app.specs[0].Matrix = windowLayerMatrix(30)
	shell.drawReplayPartial(dst)
	if shell.partial.mode != repaintRect {
		t.Fatalf("move mode = %v, want a rect repaint", shell.partial.mode)
	}
	if !covers(shell.partial.last, image.Rect(10, 10, 30, 30)) {
		t.Fatalf("move damage %v does not cover the old pose", shell.partial.last)
	}
	if !covers(shell.partial.last, image.Rect(40, 10, 60, 30)) {
		t.Fatalf("move damage %v does not cover the new pose", shell.partial.last)
	}

	app.specs = nil
	shell.drawReplayPartial(dst)
	if shell.renderer.LayersActive(d) {
		t.Fatal("a dropped spec left the layer active")
	}
	if shell.partial.mode != repaintRect {
		t.Fatalf("demotion mode = %v, want a rect repaint", shell.partial.mode)
	}
}

// covers reports whether one of rects contains want.
func covers(rects []image.Rectangle, want image.Rectangle) bool {
	for _, rect := range rects {
		if want.In(rect) {
			return true
		}
	}

	return false
}

func TestScrollBufferRespectsFixedPaintOrder(t *testing.T) {
	s := &shell{app: &fakeScreen{}, display: &layout.Display{
		Ops:   []layout.DisplayOp{{Kind: layout.DisplayOpFillRect, Fixed: true}, {Kind: layout.DisplayOpFillRect}},
		Order: []int{0, 1},
	}}
	if s.bufferOrderSafe() {
		t.Fatal("scroll buffer would cover higher content with lower fixed layer")
	}
	s.display = &layout.Display{Ops: s.display.Ops, Order: []int{1, 0}}
	if !s.bufferOrderSafe() {
		t.Fatal("top viewport layer can use cached scroll buffer")
	}
}

// An edit below the canvas must partially repaint the content-sized buffer:
// the buffer covers the overflow, and the dirty rect must reach DrawRect.
func TestOverflowEditRepaintsOneRect(t *testing.T) {
	t.Parallel()

	const ppt = 72.0 / 96.0
	boxes := []layout.Box{{X: 0, Y: 0, W: 100, H: 300}}
	display := &layout.Display{
		Width: 100, Height: 100, PointsPerPixel: ppt,
		Boxes: boxes,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 75, H: 75, R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 112.5, W: 75, H: 15, R: 1, Alpha: 1},
		},
		Order: []int{0, 1},
	}

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = display
	screen.boxes = boxes
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	if s.partial.tiles == nil || s.partial.tiles.content() != image.Rect(0, 0, 100, 300) {
		t.Fatalf("grid covers %v, want the 100x300 content", s.partial.tiles)
	}

	overflow := image.Rect(0, 140, 100, 180)
	screen.click(overflow)
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != overflow {
		t.Fatalf("mode = %v, rects = %v, want a repaint of %v", s.partial.mode, s.partial.last, overflow)
	}
}

func TestClickRepaintsOneRect(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	screen.click(image.Rect(10, 10, 30, 30))
	s.drawReplayPartial(dst)

	if screen.layouts != 1 {
		t.Fatalf("layouts = %d, want 1", screen.layouts)
	}

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != image.Rect(10, 10, 30, 30) {
		t.Fatalf("mode = %v, rects = %v", s.partial.mode, s.partial.last)
	}

	screen.ok = false
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintBlit {
		t.Fatalf("idle mode = %v, want blit", s.partial.mode)
	}

	if screen.takes != 3 {
		t.Fatalf("damage calls = %d, want 3", screen.takes)
	}
}

func TestRemovedOpRepaintsOldBounds(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	// The box vanishes. The page's diff reports its old bounds, and the
	// window must repaint exactly those so the old pixels are cleared.
	gone := partialDisplay(100, 100)
	gone.Ops = gone.Ops[:1]
	gone.Order = []int{0}
	screen.display, s.display = gone, gone
	screen.click(image.Rect(9, 9, 31, 31))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != image.Rect(9, 9, 31, 31) {
		t.Fatalf("mode = %v, rects = %v", s.partial.mode, s.partial.last)
	}
}

func TestPinnedChromeKeepsTilePath(t *testing.T) {
	t.Parallel()

	screen, s := newPinShell()
	dst := partialBuf(100, 200)

	s.drawReplayPartial(dst)

	if !s.hasViewportOps() {
		t.Fatal("fixture has no viewport ops")
	}

	if s.fullReplay() {
		t.Fatal("hasViewportOps forced fullReplay")
	}

	if s.partial.tiles == nil || s.partial.paints == 0 || s.partial.pins == 0 {
		t.Fatal("pinned chrome left the tile path")
	}

	screen.click(image.Rect(10, 80, 90, 120))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || s.partial.tiles == nil || s.partial.pins == 0 {
		t.Fatalf("message dirty: mode = %v pins = %d", s.partial.mode, s.partial.pins)
	}
}

func TestPinnedChromeIdleOnBlit(t *testing.T) {
	t.Parallel()

	screen, s := newPinShell()
	dst := partialBuf(100, 200)

	s.drawReplayPartial(dst)
	s.frameCommit()
	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()

	if !s.hasViewportOps() || s.fullReplay() {
		t.Fatal("idle pin frame must not full-replay")
	}

	if s.partial.pins == 0 {
		t.Fatal("idle blit skipped pinned chrome")
	}

	if !s.idleCleanFrame() {
		t.Fatal("valid tiles with viewport ops must idle")
	}
}

// TestPinnedChromePartialFrameRestoresVisibleTiles guards against translucent
// fixed/pinned layers (such as modal backdrops) double-blending over their own
// retained pixels on partial or idle frames: only the damaged tile repaints,
// while the visible tiles blit cleanly before pinned chrome paints on top.
func TestPinnedChromePartialFrameRestoresVisibleTiles(t *testing.T) {
	t.Parallel()

	d := &layout.Display{
		Width: 400, Height: 400, PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 400, H: 400, R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 400, H: 400, Alpha: 0.38, Fixed: true, ZIndex: 90, ZIndexSet: true},
		},
		Order: []int{0, 1, 2},
	}
	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 400, height: 400}, display: d, gen: 1}
	s := newDirtyShell(screen)
	dst := partialBuf(400, 400)
	defer dst.Deallocate()

	s.drawReplayPartial(dst)
	s.frameCommit()

	// Partial damage in tile (0,0) repaints only the damaged rect while still
	// restoring the entire 400x400 viewport from cached tiles before DrawFixed.
	screen.click(image.Rect(10, 10, 30, 30))
	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.partial.mode != repaintRect {
		t.Fatalf("mode = %v, want repaintRect", s.partial.mode)
	}
	if s.partial.paintPx != 400 {
		t.Fatalf("paintPx = %d, want 400 (only the 20x20 damaged rect)", s.partial.paintPx)
	}
	if s.partial.blitPx != 400*400 {
		t.Fatalf("blitPx = %d, want %d (full visible tile restore)", s.partial.blitPx, 400*400)
	}
	if s.partial.pins == 0 {
		t.Fatal("partial frame skipped pinned chrome")
	}

	// Idle frame repaints zero tiles, still restores visible tiles before
	// DrawFixed, and remains idle-clean.
	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.partial.paints != 0 {
		t.Fatalf("idle paints = %d, want 0", s.partial.paints)
	}
	if s.partial.blitPx != 400*400 {
		t.Fatalf("idle blitPx = %d, want %d", s.partial.blitPx, 400*400)
	}
	if !s.idleCleanFrame() {
		t.Fatal("idle frame with translucent fixed overlay must stay idle-clean")
	}

	// Short content (smaller than viewport) falls back to showTilesWhole so
	// uncovered margins below the content are cleared before DrawFixed.
	short := &layout.Display{
		Width: 400, Height: 120, PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 400, H: 120, R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 400, H: 400, Alpha: 0.38, Fixed: true, ZIndex: 90, ZIndexSet: true},
		},
		Order: []int{0, 1},
	}
	shortScreen := &dirtyScreen{fakeScreen: &fakeScreen{width: 400, height: 400}, display: short, gen: 1}
	shortShell := newDirtyShell(shortScreen)
	shortShell.drawReplayPartial(dst)
	shortShell.frameCommit()
	shortScreen.click(image.Rect(10, 10, 30, 30))
	shortShell.drawReplayPartial(dst)
	shortShell.frameCommit()
	if shortShell.partial.blitPx != 400*120 {
		t.Fatalf("short page blitPx = %d, want %d", shortShell.partial.blitPx, 400*120)
	}
}

func TestDockClipsPinnedChrome(t *testing.T) {
	t.Parallel()

	d := &layout.Display{
		Width: 600, Height: 300, PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			pinOp(0, 0, 600, 300, 1, 1, 1, 0),
			pinOp(0, 0, 600, 40, 1, 0, 0, pinLayer),
		},
		Order: []int{0, 1},
	}
	screen := &pinScreen{dirtyScreen: &dirtyScreen{
		fakeScreen: &fakeScreen{width: 600, height: 300},
		display:    d, gen: 1,
	}}
	s := newDirtyShell(screen.dirtyScreen)
	s.app, s.display = screen, d
	s.screenW, s.screenH = 600, 300
	s.dev.on = true
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)

	if s.partial.tiles == nil || !s.hasViewportOps() || s.fullReplay() {
		t.Fatal("docked pin chrome left the tile path")
	}

	clip := s.pinClip(dst)
	viewW, _ := s.viewSize()
	if clip == nil || clip.Bounds().Dx() != viewW || viewW >= 600 {
		t.Fatalf("pin clip = %v viewW = %d, want the dock strip", clip, viewW)
	}

	if s.partial.pins == 0 {
		t.Fatal("docked frame skipped pinned chrome")
	}
}

func TestScrollKeepsPinnedChrome(t *testing.T) {
	t.Parallel()

	screen, s := newPinShell()
	dst := partialBuf(100, 200)

	s.drawReplayPartial(dst)
	grid := s.partial.tiles
	if grid == nil {
		t.Fatal("no grid")
	}

	s.scrollY = 40
	screen.ok = false
	s.drawReplayPartial(dst)

	if s.partial.tiles != grid {
		t.Fatal("scroll rebuilt the grid")
	}

	if s.partial.paints != 0 || !s.partial.whole {
		t.Fatalf("scroll paints = %d whole = %v, want blit", s.partial.paints, s.partial.whole)
	}

	if s.partial.pins == 0 {
		t.Fatal("scroll skipped pinned chrome")
	}

	if s.fullReplay() || !s.hasViewportOps() {
		t.Fatal("scroll with pins must stay on the tile path")
	}
}

func TestPlanRepaintsPicksTheWork(t *testing.T) {
	t.Parallel()

	state := partialState{tiles: newTileGrid(100, 100), gen: 1, clipX: 100}
	badSize := state
	badSize.tiles = newTileGrid(80, 100)
	full := image.Rect(0, 0, 100, 100)
	rebuild := []repaintPlan{{mode: repaintBuffer, rect: full}}
	blit := []repaintPlan{{mode: repaintBlit, rect: full}}
	one := []image.Rectangle{image.Rect(10, 10, 20, 20)}
	two := []image.Rectangle{one[0], image.Rect(60, 60, 80, 80)}
	overflow := make([]image.Rectangle, maxRepaintRects+1)
	for i := range overflow {
		overflow[i] = image.Rect(i*10, 0, i*10+5, 5)
	}

	cases := []struct {
		name  string
		st    partialState
		gen   uint64
		rects []image.Rectangle
		ok    bool
		want  []repaintPlan
	}{
		{"first frame", partialState{}, 1, one, true, rebuild},
		{"one rect", state, 2, one, true,
			[]repaintPlan{{mode: repaintRect, rect: one[0]}}},
		{"two rects", state, 2, two, true,
			[]repaintPlan{{mode: repaintRect, rect: two[0]}, {mode: repaintRect, rect: two[1]}}},
		{"full rect", state, 2, []image.Rectangle{full}, true, rebuild},
		{"overflow", state, 2, overflow, true, rebuild},
		{"empty", state, 1, nil, true, blit},
		{"idle", state, 1, nil, false, blit},
		{"stale idle", state, 2, nil, false, rebuild},
		{"size change", badSize, 2, one, true, rebuild},
		{"clamped", state, 2, []image.Rectangle{image.Rect(-20, -20, 30, 30)}, true,
			[]repaintPlan{{mode: repaintRect, rect: image.Rect(0, 0, 30, 30)}}},
		{"away", state, 2, []image.Rectangle{image.Rect(-30, -30, -20, -20)}, true, rebuild},
	}

	for _, c := range cases {
		got := planRepaints(100, 100, c.st, c.gen, c.rects, c.ok, 100)
		if !slices.Equal(got, c.want) {
			t.Fatalf("%s: plans = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestPlanRepaintsClipsToTheDock checks the dock clip in the planner: hidden
// damage plans nothing, a crossing rect trims, and a grown clip repaints the
// strip the dock hid.
func TestPlanRepaintsClipsToTheDock(t *testing.T) {
	t.Parallel()

	state := partialState{tiles: newTileGrid(100, 100), gen: 1, clipX: 100}
	opened := state
	opened.clipX = 50
	blit := []repaintPlan{{mode: repaintBlit, rect: image.Rect(0, 0, 100, 100)}}
	one := []image.Rectangle{image.Rect(10, 10, 20, 20)}
	strip := repaintPlan{mode: repaintRect, rect: image.Rect(50, 0, 100, 100)}

	cases := []struct {
		name  string
		st    partialState
		gen   uint64
		rects []image.Rectangle
		ok    bool
		clip  int
		want  []repaintPlan
	}{
		{"hidden", state, 2, []image.Rectangle{image.Rect(60, 60, 80, 80)}, true, 50, blit},
		{"clipped", state, 2, []image.Rectangle{image.Rect(20, 20, 40, 40)}, true, 30,
			[]repaintPlan{{mode: repaintRect, rect: image.Rect(20, 20, 30, 40)}}},
		{"strip after close", opened, 1, nil, false, 100, []repaintPlan{strip}},
		{"strip joins damage", opened, 2, one, true, 100,
			[]repaintPlan{{mode: repaintRect, rect: one[0]}, strip}},
	}

	for _, c := range cases {
		got := planRepaints(100, 100, c.st, c.gen, c.rects, c.ok, c.clip)
		if !slices.Equal(got, c.want) {
			t.Fatalf("%s: plans = %v, want %v", c.name, got, c.want)
		}
	}
}

// windowingScreen is a dirtyScreen that observes scroll offsets, the way a
// row-windowed page does.
type windowingScreen struct {
	*dirtyScreen
	ox, oy    int
	windowing bool
}

func (w *windowingScreen) SetScrollOffset(x, y int) { w.ox, w.oy = x, y }
func (w *windowingScreen) ScrollOffset() (int, int) { return w.ox, w.oy }
func (w *windowingScreen) Windowing() bool          { return w.windowing }

func newWindowingShell(inner *dirtyScreen, on bool) *shell {
	app := &windowingScreen{dirtyScreen: inner, windowing: on}

	return &shell{
		app: app, ctx: context.Background(),
		display: inner.display, seq: inner.gen,
		screenW: inner.width, screenH: inner.height,
	}
}

func windowingDisplay(w, h int) *dirtyScreen {
	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 800, height: 600}}
	screen.display = partialDisplay(w, h)
	screen.gen = 1

	return screen
}

// TestWindowingRoutesToViewport checks that a page with the windowing opt-in
// repaints through the viewport-sized buffer even when its content fits the
// tile budget, while the same content without the opt-in keeps the tile grid.
func TestWindowingRoutesToViewport(t *testing.T) {
	t.Parallel()

	plain := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	plain.display = partialDisplay(1920, 2200)
	plain.gen = 1
	s := newDirtyShell(plain)
	dst := partialBuf(1920, 1080)
	defer dst.Deallocate()
	s.drawReplayPartial(dst)
	if s.viewport.buf != nil {
		t.Fatal("an ordinary page took the viewport path")
	}
	if s.partial.tiles == nil {
		t.Fatal("no content grid")
	}

	inner := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	inner.display = partialDisplay(1920, 2200)
	inner.gen = 1
	w := newWindowingShell(inner, true)
	wdst := partialBuf(1920, 1080)
	defer wdst.Deallocate()
	w.drawReplayPartial(wdst)
	if w.viewport.buf == nil {
		t.Fatal("a windowing page kept the content grid")
	}
	if w.partial.tiles != nil {
		t.Fatal("the content grid was retained")
	}
	if got := w.viewport.buf.Bounds().Size(); got != image.Pt(1920, 1080+viewportSides*viewportOverscan) {
		t.Fatalf("viewport buffer = %v, want the window plus overscan", got)
	}
}

// TestWindowingOptOutKeepsTilePath checks that a page carrying the scroll
// observer role but opted out takes exactly the old path.
func TestWindowingOptOutKeepsTilePath(t *testing.T) {
	t.Parallel()

	inner := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	inner.display = partialDisplay(1920, 2200)
	inner.gen = 1
	s := newWindowingShell(inner, false)
	dst := partialBuf(1920, 1080)
	defer dst.Deallocate()
	s.drawReplayPartial(dst)
	if s.viewport.buf != nil {
		t.Fatal("an opted-out page took the viewport path")
	}
	if s.partial.tiles == nil {
		t.Fatal("no content grid")
	}
}

// TestWindowingKeepsScrollRange checks that the opt-in changes only the
// repaint working set: the scrollable content size, which carries the full
// spacer height, is identical with and without windowing.
func TestWindowingKeepsScrollRange(t *testing.T) {
	t.Parallel()

	plain := windowingDisplay(800, 40200)
	plain.boxes = []layout.Box{{W: 800, H: 40200}}
	s := newDirtyShell(plain)
	plainW, plainH := s.contentSize()

	inner := windowingDisplay(800, 40200)
	inner.boxes = []layout.Box{{W: 800, H: 40200}}
	w := newWindowingShell(inner, true)
	winW, winH := w.contentSize()

	if plainW != winW || plainH != winH {
		t.Fatalf("content = %dx%d windowing, %dx%d ordinary", winW, winH, plainW, plainH)
	}
	if winW != 800 || winH != 40200 {
		t.Fatalf("content = %dx%d, want the full spacer height", winW, winH)
	}
}

// TestWindowingWarmupBoundedToViewport checks that the cache-warming sweep
// for a windowed page covers the viewport plus overscan instead of the full
// spacer height.
func TestWindowingWarmupBoundedToViewport(t *testing.T) {
	t.Parallel()

	var warmup viewportWarmup
	content, view := image.Pt(800, 40200), image.Pt(800, 600)
	warmup.take(content, view, image.Point{}, 1, true)
	if warmup.area.Empty() {
		t.Fatal("no windowing warmup area")
	}
	if want := view.Y + viewportSides*viewportOverscan; warmup.area.Dy() != want {
		t.Fatalf("windowing warmup height = %d, want %d", warmup.area.Dy(), want)
	}
	if warmup.area.Dx() != content.X || warmup.area.Min.X != 0 {
		t.Fatalf("windowing warmup width = %v, want the full content width", warmup.area)
	}
	if r := warmup.take(content, view, image.Point{}, 1, true); r.Empty() || r.Dy() > prewarmBandHeight {
		t.Fatalf("windowing warmup band = %v, want one capped band", r)
	}

	scroll := image.Pt(0, 30000)
	warmup.take(content, view, scroll, 1, true)
	wanted := image.Rectangle{Min: scroll, Max: scroll.Add(view)}
	if !wanted.In(warmup.area) {
		t.Fatalf("warmup did not follow the viewport: area=%v", warmup.area)
	}
	if warmup.area.Dy() > view.Y+viewportSides*viewportOverscan {
		t.Fatalf("followed warmup height = %d, want at most the viewport plus overscan", warmup.area.Dy())
	}
}

// TestWindowingWarmupStaysViewportSized drives one viewport frame for a
// compare-like page and checks the sweep the frame scheduled, while the same
// page without the opt-in keeps the old twelve-megapixel neighbourhood.
func TestWindowingWarmupStaysViewportSized(t *testing.T) {
	t.Parallel()

	inner := windowingDisplay(800, 40200)
	w := newWindowingShell(inner, true)
	dst := partialBuf(800, 600)
	defer dst.Deallocate()
	w.drawReplayPartial(dst)
	if got := w.viewport.warmup.area.Dy(); got != 600+viewportSides*viewportOverscan {
		t.Fatalf("windowing sweep height = %d, want the viewport plus overscan", got)
	}

	plain := windowingDisplay(800, 40200)
	s := newDirtyShell(plain)
	sdst := partialBuf(800, 600)
	defer sdst.Deallocate()
	s.drawReplayPartial(sdst)
	if got := s.viewport.warmup.area.Dy(); got != prewarmMaxPixels/800 {
		t.Fatalf("ordinary sweep height = %d, want the old bounded neighbourhood", got)
	}
}

func TestPaintSurfaceTransitions(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{
		fakeScreen: &fakeScreen{width: 100, height: 100},
		display:    partialDisplay(100, 100),
		gen:        1,
	}
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)
	t.Cleanup(dst.Deallocate)
	t.Cleanup(func() { s.usePaintSurface(surfaceUnset) })
	draw := func(mode paintSurface, clean bool) {
		t.Helper()
		s.drawContent(dst)
		s.frameCommit()
		if s.surface != mode || s.idleCleanFrame() != clean {
			t.Fatalf("surface = %v, clean = %t; want %v, %t", s.surface, s.idleCleanFrame(), mode, clean)
		}
	}

	draw(surfaceTiles, false)
	draw(surfaceTiles, true)
	s.pageZoom = 1.25
	draw(surfaceScaled, false)
	if s.partial.tiles != nil || s.scaleBuf == nil {
		t.Fatal("zoom should release tiles and use a scaling texture")
	}
	s.zoomReset()
	draw(surfaceTiles, false)
	if s.partial.paints == 0 || s.scaleBuf != nil {
		t.Fatal("return to tiles should rebuild the grid and release the scaling texture")
	}

	// A screen without damage reporting replays directly. Its old tiles
	// cannot supply the first frame once damage reporting resumes.
	s.app = screen.fakeScreen
	draw(surfaceDirect, false)
	s.app = screen
	draw(surfaceTiles, false)
	if s.partial.paints == 0 {
		t.Fatal("return from direct replay should rebuild the grid")
	}

	screen.display.Ops[1].SetBlendGroup(&layout.DisplayGroup{BackdropFilter: "blur(8px)"})
	draw(surfaceBackdrop, false)
	backdrop := s.scaleBuf
	draw(surfaceBackdrop, true)
	if s.scaleBuf != backdrop {
		t.Fatal("unchanged backdrop should keep its painted texture")
	}

	// A zoom draw has the same texture dimensions as the backdrop in this
	// fixture. The next backdrop frame must paint, even with the same display,
	// generation, scroll and no dirty rectangles.
	s.pageZoom = 1.25
	draw(surfaceScaled, false)
	s.zoomReset()
	draw(surfaceBackdrop, false)
	draw(surfaceBackdrop, true)

	s.display = nil
	draw(surfaceBitmap, false)
	if s.scaleBuf != nil || s.partial.backdropDisplay != nil {
		t.Fatal("a frame without an image or display must release the old backdrop")
	}
}
