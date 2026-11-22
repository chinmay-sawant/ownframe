package window

import (
	"bytes"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"os"
	"testing"
)

// TestRebuildCoversEveryTile checks the first frame: the grid is built for
// the content and every tile is painted.
func TestRebuildCoversEveryTile(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)

	s.drawReplayPartial(partialBuf(600, 300))

	if s.partial.mode != repaintBuffer {
		t.Fatalf("mode = %v, want a rebuild", s.partial.mode)
	}

	g := s.partial.tiles
	if g == nil || g.cols != 3 || g.rows != 2 {
		t.Fatalf("grid = %+v, want 3x2 tiles", g)
	}

	if s.partial.paints != g.cols*g.rows {
		t.Fatalf("painted %d tiles, want %d", s.partial.paints, g.cols*g.rows)
	}

	if s.partial.buf == nil || !s.partial.buf.Bounds().Empty() {
		t.Fatal("the retained-content handle should be an empty marker")
	}

	if len(s.partial.last) != 1 || s.partial.last[0] != image.Rect(0, 0, 600, 300) {
		t.Fatalf("last = %v, want the full content", s.partial.last)
	}
}

// TestDirtyRectPaintsOnlyItsTiles checks a click repaints the tiles its box
// meets and no others.
func TestDirtyRectPaintsOnlyItsTiles(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)

	screen.click(image.Rect(10, 10, 20, 20))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || s.partial.paints != 1 {
		t.Fatalf("small rect: mode = %v, paints = %d, want one tile",
			s.partial.mode, s.partial.paints)
	}

	cross := image.Rect(250, 250, 300, 300)
	screen.click(cross)
	s.drawReplayPartial(dst)

	if s.partial.paints != 4 {
		t.Fatalf("across tiles: painted %d tiles, want 4", s.partial.paints)
	}

	if len(s.partial.last) != 1 || s.partial.last[0] != cross {
		t.Fatalf("last = %v, want the click rect", s.partial.last)
	}
}

// TestTileGridKeepsTheContentCost checks the grid partitions the content:
// every tile is clipped at the edge and the tiles add up to the content area,
// the single buffer's cost.
func TestTileGridKeepsTheContentCost(t *testing.T) {
	t.Parallel()

	sizes := []image.Point{{100, 100}, {256, 256}, {600, 500}, {1908, 999}}

	for _, size := range sizes {
		g := newTileGrid(size.X, size.Y)
		if g == nil {
			t.Fatalf("%v: no grid", size)
		}

		want := gridTiles(size.X) * gridTiles(size.Y)
		if len(g.images) != want {
			t.Fatalf("%v: %d tiles, want %d", size, len(g.images), want)
		}

		area := 0

		for row := 0; row < g.rows; row++ {
			for col := 0; col < g.cols; col++ {
				r := g.rectAt(col, row)
				if r.Dx() > tileSize || r.Dy() > tileSize || !r.In(g.content()) {
					t.Fatalf("%v: tile %d,%d = %v", size, col, row, r)
				}

				if b := g.at(col, row).Bounds(); b != image.Rect(0, 0, r.Dx(), r.Dy()) {
					t.Fatalf("%v: tile %d,%d image = %v, want %v", size, col, row, b, r)
				}

				area += r.Dx() * r.Dy()
			}
		}

		if area != size.X*size.Y {
			t.Fatalf("%v: tiles cover %d pixels, want %d", size, area, size.X*size.Y)
		}
	}
}

// TestEdgeTileIsClipped pins the corner tile of a size that is not a
// multiple of the tile size: it is the remainder, not a full tile.
func TestEdgeTileIsClipped(t *testing.T) {
	t.Parallel()

	g := newTileGrid(300, 300) // 2 columns, 2 rows
	if got := g.rectAt(1, 1); got != image.Rect(256, 256, 300, 300) {
		t.Fatalf("corner tile = %v, want a 44x44 corner", got)
	}

	if b := g.at(1, 1).Bounds(); b != image.Rect(0, 0, 44, 44) {
		t.Fatalf("corner image = %v, want 44x44", b)
	}

	// A tile beside the edge keeps the full tile size: only the far edges
	// clip.
	full := newTileGrid(600, 600)
	if got := full.rectAt(0, 0); got != image.Rect(0, 0, 256, 256) || full.at(1, 1).Bounds() != image.Rect(0, 0, 256, 256) {
		t.Fatalf("interior tile = %v, want a full tile", got)
	}
}

// TestDroppedHandleForcesARebuild pins the retained-content handle contract:
// a caller that drops it (a screenshot, a GPU reset, the gated frame test)
// invalidates the tiles, so the next frame rebuilds them instead of blitting
// pixels nobody can vouch for.
func TestDroppedHandleForcesARebuild(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	screen.ok = false
	before := s.partial.tiles

	if s.partial.buf == nil {
		t.Fatal("no retained-content handle after a rebuild")
	}

	s.partial.buf = nil
	s.drawReplayPartial(dst)

	if s.partial.tiles == nil || s.partial.tiles == before {
		t.Fatal("dropped handle did not rebuild the grid")
	}

	if s.partial.mode != repaintBuffer {
		t.Fatalf("mode = %v, want a rebuild", s.partial.mode)
	}

	if s.partial.paints != s.partial.tiles.cols*s.partial.tiles.rows {
		t.Fatalf("painted %d tiles, want the whole new grid", s.partial.paints)
	}
}

// TestTileRangeMapsADirtyRect pins the tile mapping: a rect maps to the
// range of tiles it intersects, clamped to the content.
func TestTileRangeMapsADirtyRect(t *testing.T) {
	t.Parallel()

	g := newTileGrid(600, 500) // 3 columns, 2 rows

	cases := []struct {
		rect image.Rectangle
		want [4]int
	}{
		{image.Rect(10, 10, 20, 20), [4]int{0, 0, 1, 1}},
		{image.Rect(250, 250, 300, 300), [4]int{0, 0, 2, 2}},
		{image.Rect(255, 10, 257, 20), [4]int{0, 0, 2, 1}},
		{image.Rect(0, 0, 600, 500), [4]int{0, 0, 3, 2}},
		{image.Rect(-40, -40, 10, 10), [4]int{0, 0, 1, 1}},
		{image.Rect(512, 256, 600, 500), [4]int{2, 1, 3, 2}},
		{image.Rect(600, 500, 700, 700), [4]int{0, 0, 0, 0}},
	}

	for _, c := range cases {
		c0, r0, c1, r1 := g.tileRange(c.rect)
		if got := [4]int{c0, r0, c1, r1}; got != c.want {
			t.Fatalf("tileRange(%v) = %v, want %v", c.rect, got, c.want)
		}
	}
}

// TestContentChangeRebuildsTheGrid checks a content size change disposes the
// old grid and allocates one for the new content. Only the visible tiles
// paint; the row below the fold stays dirty until a scroll brings it on
// screen, when the blit repaints it on demand.
func TestContentChangeRebuildsTheGrid(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)

	s.drawReplayPartial(partialBuf(600, 300))
	first := s.partial.tiles

	screen.boxes = []layout.Box{{X: 0, Y: 0, W: 600, H: 700}}
	screen.click(image.Rect(10, 10, 20, 20))
	s.drawReplayPartial(partialBuf(600, 300))

	if s.partial.tiles == nil || s.partial.tiles == first {
		t.Fatal("size change kept the grid")
	}

	if got := s.partial.tiles.content(); got != image.Rect(0, 0, 600, 700) {
		t.Fatalf("new grid covers %v, want the 600x700 content", got)
	}

	if s.partial.paints != 6 {
		t.Fatalf("painted %d tiles, want the 6 visible ones", s.partial.paints)
	}

	s.scrollY = 400
	screen.ok = false
	s.drawReplayPartial(partialBuf(600, 300))

	if s.partial.paints != 3 {
		t.Fatalf("scroll painted %d tiles, want the 3 newly visible ones", s.partial.paints)
	}
}

// TestScrollMovesBlitsWithoutRepainting checks a scroll frame repaints no
// tile and copies the whole content at the new offset instead.
func TestScrollMovesBlitsWithoutRepainting(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(600, 300)

	s.Draw(dst)

	grid := s.partial.tiles
	if grid == nil {
		t.Fatal("no grid after the first frame")
	}

	s.scrollY = 20
	screen.ok = false
	s.Draw(dst)

	if s.partial.mode != repaintBlit {
		t.Fatalf("scroll mode = %v, want blit", s.partial.mode)
	}

	if !s.partial.whole {
		t.Fatal("scroll frame did not mark the whole content")
	}

	if s.partial.paints != 0 {
		t.Fatalf("scroll repainted %d tiles, want 0", s.partial.paints)
	}

	if s.partial.tiles != grid {
		t.Fatal("scroll rebuilt the grid")
	}

	if full, rects := blitPlan(s.partial.plans, s.partial.chrome); full || len(rects) != 0 {
		t.Fatalf("blit plan = %v (full = %v), want the whole copy", rects, full)
	}
}

// TestTileRangeTouchesExactlyTheTiles checks the mapping both ways: every
// tile in the span meets the rect, and no tile outside it does.
func TestTileRangeTouchesExactlyTheTiles(t *testing.T) {
	t.Parallel()

	g := newTileGrid(500, 300) // 2 columns, 2 rows
	rects := []image.Rectangle{
		image.Rect(10, 10, 20, 20),
		image.Rect(240, 140, 280, 180),
		image.Rect(0, 0, 500, 300),
		image.Rect(-50, -50, 5, 5),
		image.Rect(499, 299, 520, 320),
	}

	for _, rect := range rects {
		c0, r0, c1, r1 := g.tileRange(rect)

		for row := 0; row < g.rows; row++ {
			for col := 0; col < g.cols; col++ {
				in := col >= c0 && col < c1 && row >= r0 && row < r1
				if meets := !rect.Intersect(g.rectAt(col, row)).Empty(); in != meets {
					t.Fatalf("rect %v tile %d,%d: in span = %v, meets = %v", rect, col, row, in, meets)
				}
			}
		}
	}
}

// TestWithinTilesNeedsTheChromeInside pins the chrome-restore guard: chrome
// that reaches past the content forces a full frame.
func TestWithinTilesNeedsTheChromeInside(t *testing.T) {
	t.Parallel()

	g := newTileGrid(100, 100)
	if !withinTiles(g, []image.Rectangle{image.Rect(0, 0, 100, 100)}) {
		t.Fatal("a rect inside the content was rejected")
	}

	if withinTiles(g, []image.Rectangle{image.Rect(0, 0, 101, 10)}) {
		t.Fatal("chrome past the content was accepted")
	}
}

// viewportTaker reports scripted dirty rects for the oversized path.
type viewportTaker struct {
	rects []image.Rectangle
	ok    bool
	takes int
}

func (v *viewportTaker) TakeDirtyRects() ([]image.Rectangle, bool) {
	v.takes++

	return v.rects, v.ok
}

func TestViewportBufferStaysWindowSized(t *testing.T) {
	app := &dirtyScreen{fakeScreen: &fakeScreen{width: 80, height: 60}, gen: 1}
	app.display = partialDisplay(80, 37000)
	s := newDirtyShell(app)
	dst := partialBuf(80, 60)
	s.partial.tiles = newTileGrid(80, 100)
	s.drawContent(dst)
	if got := s.viewport.buf.Bounds(); got != image.Rect(0, 0, 80, 180) {
		t.Fatalf("buffer = %v", got)
	}
	if s.partial.tiles != nil {
		t.Fatal("content grid retained")
	}
	saved := s.viewport.buf
	s.drawContent(dst)
	if s.viewport.buf != saved {
		t.Fatal("idle frame reallocated")
	}
	s.scrollY = 2000
	s.drawContent(dst)
	if s.viewport.area.Min.Y > 2000 || s.viewport.area.Max.Y < 2060 || s.viewport.buf != saved {
		t.Fatal("scroll lost or replaced the buffer")
	}
	app.gen++
	s.seq = app.gen
	s.drawContent(dst)
	if s.viewport.key.generation != app.gen {
		t.Fatal("generation did not invalidate")
	}
	if app.takes != 4 {
		t.Fatalf("takes = %d, want 4", app.takes)
	}
	s.disposeViewport()
	if s.viewport.buf != nil || s.viewport.valid {
		t.Fatal("viewport not released")
	}
}

func TestViewportDirtyBounds(t *testing.T) {
	one := image.Rect(10, 20, 30, 40)
	if got := viewportDirtyBounds([]image.Rectangle{one}); got != one {
		t.Fatalf("one rect = %v, want %v", got, one)
	}
	two := []image.Rectangle{one, image.Rect(200, 300, 300, 400)}
	if got := viewportDirtyBounds(two); got != image.Rect(10, 20, 300, 400) {
		t.Fatalf("two rects = %v", got)
	}
	if got := viewportDirtyBounds(nil); !got.Empty() {
		t.Fatalf("empty list = %v, want empty", got)
	}
}

// promotedLayerScreen records the layer ids the window hands back to the page.
type promotedLayerScreen struct {
	*layerDirtyScreen
	promoted []string
}

func (s *promotedLayerScreen) SetPromotedLayers(ids []string) {
	s.promoted = append([]string(nil), ids...)
}

// TestViewportCompositesRetainedLayer pins the background-plus-layer path: a
// tall page bakes the promoted card into a layer, leaves its operations out of
// the cached background, and reports the promoted id back to the page so the
// next sample stops damaging the page under it.
func TestViewportCompositesRetainedLayer(t *testing.T) {
	d := windowLayerDisplay()
	app := &promotedLayerScreen{layerDirtyScreen: &layerDirtyScreen{
		dirtyScreen: &dirtyScreen{
			fakeScreen: &fakeScreen{width: 100, height: 100},
			display:    d,
			gen:        1,
			ok:         true,
		},
		specs: []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: windowLayerMatrix(2), Alpha: 1}},
	}}
	shell := newDirtyShell(app.dirtyScreen)
	shell.app = app
	shell.display = d
	dst := ebiten.NewImage(100, 100)
	defer dst.Deallocate()

	shell.drawViewport(dst, d, &viewportTaker{})

	if shell.viewport.buf == nil {
		t.Fatal("the viewport path kept no background buffer")
	}
	if !shell.renderer.LayersActive(d) {
		t.Fatal("the card was not promoted")
	}
	if len(app.promoted) != 1 || app.promoted[0] != "card" {
		t.Fatalf("promoted ids = %v, want [card]", app.promoted)
	}
}

func TestViewportSpareRetainedAcrossScrollPause(t *testing.T) {
	app := &dirtyScreen{fakeScreen: &fakeScreen{width: 80, height: 60}, gen: 1}
	app.display = partialDisplay(80, 37000)
	s := newDirtyShell(app)
	taker := &viewportTaker{}
	dst := partialBuf(80, 60)
	s.drawViewport(dst, app.display, taker)

	// Scroll crosses overscan, triggering moveViewport which allocates spare.
	s.scrollY = 150
	s.drawViewport(dst, app.display, taker)
	if s.viewport.spare == nil {
		t.Fatal("moveViewport did not allocate spare")
	}
	spare := s.viewport.spare

	// Settle warmup and pause scrolling. Next frame is blitOnly.
	s.viewport.warmup.next.Y = s.viewport.warmup.area.Max.Y
	s.drawViewport(dst, app.display, taker)
	if !s.viewport.blitOnly {
		t.Fatal("stationary frame was not blitOnly")
	}
	if s.viewport.spare == nil || s.viewport.spare != spare {
		t.Fatal("spare was deallocated or replaced on scroll pause")
	}

	// Resuming scroll uses the retained spare.
	s.scrollY = 220
	s.drawViewport(dst, app.display, taker)
	if s.viewport.spare == nil {
		t.Fatal("spare missing after resuming scroll")
	}

	// Leaving the viewport also deallocates its spare.
	s.usePaintSurface(surfaceUnset)
	if s.viewport.spare != nil || s.viewport.buf != nil {
		t.Fatal("usePaintSurface did not release viewport buffers")
	}
}

// This isolated 100px fixture has no DevTools and never closes a dock.
func TestViewportAnimatedReturnPixels(t *testing.T) {
	runAnimatedReturnTest(t, true)
}

func TestViewportDirectAnimationReturnPixels(t *testing.T) {
	runAnimatedReturnTest(t, false)
}

func runAnimatedReturnTest(t *testing.T, promote bool) {
	t.Helper()
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check")
	}
	d := windowLayerDisplay()
	d.Height = 2000
	d.Ops[0].H = 1500
	app := &layerDirtyScreen{dirtyScreen: &dirtyScreen{
		fakeScreen: &fakeScreen{width: 100, height: 100}, display: d, gen: 1,
	}}
	s := newDirtyShell(app.dirtyScreen)
	s.app = app
	g := &animatedReturnGame{shell: s, app: app, promote: promote}
	ebiten.SetWindowSize(100, 100)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
}

type animatedReturnGame struct {
	shell   *shell
	app     *layerDirtyScreen
	step    int
	err     error
	promote bool
}

func (g *animatedReturnGame) Layout(int, int) (int, int) { return 100, 100 }
func (g *animatedReturnGame) Update() error {
	if g.err != nil || g.step == 4 {
		return ebiten.Termination
	}
	return nil
}

func (g *animatedReturnGame) Draw(dst *ebiten.Image) {
	if g.err != nil || g.step == 4 {
		return
	}
	matrix := windowLayerMatrix(15)
	if g.step >= 2 {
		matrix = windowLayerMatrix(45)
	}
	for i := 1; i < len(g.app.display.Ops); i++ {
		g.app.display.Ops[i].SetXform(matrix)
		g.app.display.Ops[i].SetPaintOpacity(0.35)
	}
	g.app.specs = nil
	g.shell.scrollY = 130
	if g.step == 0 || g.step == 3 {
		g.shell.scrollY = 0
		if g.promote {
			base, _ := replay.InvertLayerMatrix(matrix)
			g.app.specs = []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: matrix, Base: base, Alpha: 0.35}}
		}
	}
	g.shell.drawViewport(dst, g.app.display, g.app)
	expected := ebiten.NewImage(100, 100)
	defer expected.Deallocate()
	expected.Fill(g.shell.pageBackground())
	g.shell.renderer.DrawVisibleUnfixedZBackground(expected, g.app.display, 0, -float64(g.shell.scrollY), 0)
	g.shell.renderer.DrawLayers(expected, g.app.display, 0, -float64(g.shell.scrollY))
	actualPixels, expectedPixels := make([]byte, 100*100*4), make([]byte, 100*100*4)
	dst.ReadPixels(actualPixels)
	expected.ReadPixels(expectedPixels)
	if !bytes.Equal(actualPixels, expectedPixels) {
		g.err = fmt.Errorf("animation scroll stage %d differs from fresh replay", g.step)
	}
	g.step++
}

func TestViewportExposedCoversOnlyNewPixels(t *testing.T) {
	previous := image.Rect(0, 0, 100, 200)
	for _, next := range []image.Rectangle{
		image.Rect(0, 80, 100, 280), image.Rect(20, 80, 120, 280),
		image.Rect(-20, -80, 80, 120), image.Rect(0, 1000, 100, 1200), previous,
	} {
		strips := viewportExposed(next, previous)
		for y := next.Min.Y; y < next.Max.Y; y++ {
			for x := next.Min.X; x < next.Max.X; x++ {
				point := image.Pt(x, y)
				covers := 0
				for _, strip := range strips {
					if point.In(strip) {
						covers++
					}
				}
				want := 1
				if point.In(previous) {
					want = 0
				}
				if covers != want {
					t.Fatalf("%v: pixel %v replayed %d times, want %d", next, point, covers, want)
				}
			}
		}
	}
}

func TestViewportScrollPixels(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check")
	}
	const width, height = 128, 96
	d := partialDisplay(300, 2000)
	for y := 0; y < 2000; y += 80 {
		d.Ops = append(d.Ops, layout.DisplayOp{
			Kind: layout.DisplayOpFillRect, X: 15, Y: float64(y) * d.PointsPerPixel,
			W: 150, H: 24, R: 1, G: float64(y%3) / 2, Alpha: 1,
		})
		d.Order = append(d.Order, len(d.Ops)-1)
	}
	app := &dirtyScreen{fakeScreen: &fakeScreen{width: width, height: height}, display: d, gen: 1}
	g := &viewportPixelGame{shell: newDirtyShell(app), app: app}
	ebiten.SetWindowSize(width, height)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if !g.reused {
		t.Fatal("the scroll never reused the viewport overlap")
	}
}

type viewportPixelGame struct {
	shell  *shell
	app    *dirtyScreen
	step   int
	reused bool
	err    error
}

func (g *viewportPixelGame) Layout(int, int) (int, int) { return 128, 96 }
func (g *viewportPixelGame) Update() error {
	if g.err != nil || g.step == 7 {
		return ebiten.Termination
	}
	return nil
}

func (g *viewportPixelGame) Draw(dst *ebiten.Image) {
	if g.err != nil || g.step == 7 {
		return
	}
	offsets := [...]image.Point{{}, {Y: 120}, {Y: 240}, {X: 40, Y: 300}, {Y: 200}, {Y: 1400}, {}}
	at := offsets[g.step]
	g.shell.scrollX, g.shell.scrollY = at.X, at.Y
	if g.step == 2 {
		g.app.display.Ops[5].B = 1
		g.app.click(image.Rect(0, 240, 300, 280))
		g.shell.seq = g.app.gen
	} else {
		g.app.ok = false
		g.app.rects = nil
	}
	g.shell.partial.paintPx = 0
	g.shell.drawViewport(dst, g.app.display, g.app)
	if g.shell.viewport.spare != nil {
		g.reused = true
	}
	if g.step == 3 && g.shell.partial.paintPx >= 128*288 {
		g.err = fmt.Errorf("overlapping scroll repainted the full viewport: %d pixels", g.shell.partial.paintPx)
		return
	}
	expected := ebiten.NewImage(128, 96)
	defer expected.Deallocate()
	expected.Fill(g.shell.pageBackground())
	g.shell.renderer.DrawVisibleUnfixedZ(expected, g.app.display, -float64(at.X), -float64(at.Y), 0)
	actualPixels, expectedPixels := make([]byte, 128*96*4), make([]byte, 128*96*4)
	dst.ReadPixels(actualPixels)
	expected.ReadPixels(expectedPixels)
	if !bytes.Equal(actualPixels, expectedPixels) {
		g.err = fmt.Errorf("scroll stage %d at %v differs from fresh replay", g.step, at)
	}
	g.step++
}

func TestViewportWarmupEventuallyCoversThePage(t *testing.T) {
	var warmup viewportWarmup
	content, view := image.Pt(2400, 4000), image.Pt(1000, 600)
	whole := image.Rectangle{Max: content}
	var pixels int
	for {
		r := warmup.take(content, view, image.Point{}, 1, false)
		if r.Empty() {
			break
		}
		if !r.In(whole) || r.Dx() > view.X || r.Dy() > prewarmBandHeight {
			t.Fatalf("warmup exceeded the per-frame band: %v", r)
		}
		pixels += r.Dx() * r.Dy()
		if pixels > content.X*content.Y {
			t.Fatal("warmup repeated content instead of finishing")
		}
	}
	if pixels != content.X*content.Y {
		t.Fatalf("warmed %d pixels, want the whole page", pixels)
	}
	if r := warmup.take(content, view, image.Point{}, 1, false); !r.Empty() {
		t.Fatalf("completed warmup restarted: %v", r)
	}
	if r := warmup.take(content, view, image.Point{}, 2, false); r.Min != (image.Point{}) || r.Empty() {
		t.Fatalf("new content did not restart warmup: %v", r)
	}
}

func TestViewportWarmupFollowsLargePageScroll(t *testing.T) {
	var warmup viewportWarmup
	content, view := image.Pt(2000, 100000), image.Pt(1000, 600)
	warmup.take(content, view, image.Point{}, 1, false)
	if warmup.area.Dx()*warmup.area.Dy() > prewarmMaxPixels {
		t.Fatal("large-page warmup exceeds its pixel budget")
	}
	scroll := image.Pt(0, 90000)
	r := warmup.take(content, view, scroll, 1, false)
	wanted := image.Rectangle{Min: scroll, Max: scroll.Add(view)}
	if r.Empty() || !wanted.In(warmup.area) {
		t.Fatalf("warmup did not follow the viewport: band=%v area=%v", r, warmup.area)
	}
}

// TestOffscreenDirtyTileExpandsSubRectPaint pins that a tile marked dirty while
// offscreen expands a subsequent sub-rect damage repaint to the full tile
// before marking the tile clean.
func TestOffscreenDirtyTileExpandsSubRectPaint(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 256, height: 256}}
	screen.display = partialDisplay(256, 512)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(256, 256)

	// Frame 1 at scrollY=0: tile (0,0) paints and becomes clean; tile (0,1)
	// sits offscreen and stays dirty.
	s.drawReplayPartial(dst)
	g := s.partial.tiles
	if g == nil || !g.cleanAt(0, 0) || g.cleanAt(0, 1) {
		t.Fatalf("unexpected initial tile clean state: (0,0)=%v (0,1)=%v", g.cleanAt(0, 0), g.cleanAt(0, 1))
	}

	// Scroll to row 1 and send a tiny 10x10 dirty rect inside dirty tile (0,1).
	s.scrollY = 256
	s.partial.scroll = image.Pt(0, 256)
	screen.click(image.Rect(10, 266, 20, 276))
	s.drawReplayPartial(dst)

	if !g.cleanAt(0, 1) {
		t.Fatal("tile (0,1) was not marked clean after paint")
	}
	wantFullTilePx := tileSize * tileSize
	if s.partial.paintPx < wantFullTilePx {
		t.Fatalf("paintPx = %d, want at least full tile %d when repainting previously dirty tile", s.partial.paintPx, wantFullTilePx)
	}
}

// TestLayerDamageTriggersTileRepaintWhenPageDirtyEmpty verifies that tile-path
// layerDamage triggers a repaint even when TakeDirtyRects reports ok == false.
func TestLayerDamageTriggersTileRepaintWhenPageDirtyEmpty(t *testing.T) {
	t.Parallel()

	d := windowLayerDisplay()
	app := &layerDirtyScreen{
		dirtyScreen: &dirtyScreen{
			fakeScreen: &fakeScreen{width: 100, height: 100},
			display:    d,
			gen:        1,
			ok:         false,
		},
		specs: []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: windowLayerMatrix(2), Alpha: 1}},
	}
	s := newDirtyShell(app.dirtyScreen)
	s.app = app
	s.display = d
	dst := partialBuf(100, 100)

	// Frame 1: builds tile grid and promotes card.
	s.drawReplayPartial(dst)
	if !s.renderer.LayersActive(d) {
		t.Fatal("card was not promoted on tile path")
	}

	// Frame 2: page reports NO dirty rects (ok == false), but layer moved!
	app.ok = false
	app.rects = nil
	app.specs = []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: windowLayerMatrix(12), Alpha: 1}}
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect {
		t.Fatalf("mode = %v, want repaintRect when layerDamage is non-empty", s.partial.mode)
	}
	if s.partial.paints == 0 {
		t.Fatal("layerDamage did not repaint any tile when TakeDirtyRects returned ok=false")
	}
}

// TestLeavingViewportDemotesLayersAndClearsPromoted verifies that switching
// away from the viewport layer-compositing path (to tile, backdrop, scaled, or
// usePaintSurface) demotes layers and clears SetPromotedLayers.
func TestLeavingViewportDemotesLayersAndClearsPromoted(t *testing.T) {
	t.Parallel()

	makeViewportShell := func() (*shell, *promotedLayerScreen, *layout.Display, *ebiten.Image) {
		d := windowLayerDisplay()
		d.Height = 2000
		app := &promotedLayerScreen{layerDirtyScreen: &layerDirtyScreen{
			dirtyScreen: &dirtyScreen{
				fakeScreen: &fakeScreen{width: 100, height: 100},
				display:    d,
				gen:        1,
				ok:         false,
			},
			specs: []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: windowLayerMatrix(4), Alpha: 1}},
		}}
		s := newDirtyShell(app.dirtyScreen)
		s.app = app
		s.display = d
		dst := partialBuf(100, 100)
		s.drawReplayPartial(dst)
		if !s.renderer.LayersActive(d) || len(app.promoted) != 1 || app.promoted[0] != "card" {
			t.Fatalf("viewport setup failed: active=%v promoted=%v", s.renderer.LayersActive(d), app.promoted)
		}
		return s, app, d, dst
	}

	// 1. Leaving viewport releases all its promoted layers.
	s, app, d, _ := makeViewportShell()
	s.usePaintSurface(surfaceUnset)
	if s.renderer.LayersActive(d) || len(app.promoted) != 0 {
		t.Fatalf("usePaintSurface kept stale layers/promoted: active=%v promoted=%v", s.renderer.LayersActive(d), app.promoted)
	}

	// 2. Leaving viewport -> tile path with no specs.
	s, app, d, dst := makeViewportShell()
	d.Height = 100
	app.specs = nil
	app.gen++
	s.seq = app.gen
	s.drawReplayPartial(dst)
	if s.renderer.LayersActive(d) || len(app.promoted) != 0 {
		t.Fatalf("viewport->tile kept stale layers/promoted: active=%v promoted=%v", s.renderer.LayersActive(d), app.promoted)
	}

	// 3. Leaving viewport -> backdrop-filter path.
	s, app, d, dst = makeViewportShell()
	d.Ops[1].SetBlendGroup(&layout.DisplayGroup{BackdropFilter: "blur(8px)"})
	s.drawReplayPartial(dst)
	if s.renderer.LayersActive(d) || len(app.promoted) != 0 {
		t.Fatalf("viewport->backdrop kept stale layers/promoted: active=%v promoted=%v", s.renderer.LayersActive(d), app.promoted)
	}

	// 4. Leaving viewport -> zoomed scaled path via drawContent.
	s, app, d, dst = makeViewportShell()
	s.pageZoom = 1.5
	s.drawContent(dst)
	if s.renderer.LayersActive(d) || len(app.promoted) != 0 {
		t.Fatalf("viewport->zoom kept stale layers/promoted: active=%v promoted=%v", s.renderer.LayersActive(d), app.promoted)
	}
}

// TestViewportRejectsLayerCoveredByLaterStaticOp verifies end-to-end on the
// viewport path that an animated element overlapped by a later static sibling
// is not promoted above it and is not added to promoted IDs.
func TestViewportRejectsLayerCoveredByLaterStaticOp(t *testing.T) {
	t.Parallel()
	const pt = 72.0 / 96.0

	d := windowLayerDisplay()
	d.Height = 2000
	// Add a later static popover overlapping the card box (20,20)-(50,30).
	d.Ops = append(d.Ops, layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 25 * pt, Y: 18 * pt, W: 30 * pt, H: 20 * pt, G: 1, Alpha: 1,
	})
	d.Order = append(d.Order, len(d.Ops)-1)
	d.OpOwners = append(d.OpOwners, 0)

	app := &promotedLayerScreen{layerDirtyScreen: &layerDirtyScreen{
		dirtyScreen: &dirtyScreen{
			fakeScreen: &fakeScreen{width: 100, height: 100},
			display:    d,
			gen:        1,
		},
		specs: []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: windowLayerMatrix(2), Alpha: 1}},
	}}
	s := newDirtyShell(app.dirtyScreen)
	s.app = app
	s.display = d
	dst := partialBuf(100, 100)

	s.drawViewport(dst, d, &viewportTaker{})
	if s.renderer.LayersActive(d) {
		t.Fatal("viewport promoted a layer covered by a later static sibling")
	}
	if len(app.promoted) != 0 {
		t.Fatalf("promoted = %v, want empty when layer is covered by later static op", app.promoted)
	}
}
