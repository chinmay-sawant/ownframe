package window

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/clipboard"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/leakcheck"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestDisplayBackground(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Width: 320, Height: 400, PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 480, H: 600, R: 0.9, G: 0.8, B: 0.7, Alpha: 1},
		},
		Order: []int{0, 1},
	}

	c, ok := displayBackground(display)
	if !ok {
		t.Fatal("no background")
	}

	r, g, b, _ := c.RGBA()
	if r>>8 != 229 || g>>8 != 204 || b>>8 != 178 {
		t.Fatalf("color = %d,%d,%d", r>>8, g>>8, b>>8)
	}
}

func TestDisplayBackgroundMissing(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Width: 320, Height: 400, PointsPerPixel: 0.75,
		Ops:   []layout.DisplayOp{{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, Alpha: 1}},
		Order: []int{0},
	}

	if _, ok := displayBackground(display); ok {
		t.Fatal("small fill treated as background")
	}
}

func TestDisplayBackgroundExactCanvas(t *testing.T) {
	t.Parallel()

	// 320 CSS px at 0.75 px per point is a 240-point canvas. A fill that
	// spans it must be found; dividing instead of multiplying the canvas
	// conversion would reject it.
	display := &layout.Display{
		Width: 320, Height: 320, PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 240, H: 240, R: 1, G: 1, B: 1, Alpha: 1},
		},
		Order: []int{0},
	}

	if _, ok := displayBackground(display); !ok {
		t.Fatal("exact-canvas fill not found")
	}
}

func TestDisplayBackgroundPaintOrder(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Width: 320, Height: 320, PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 240, H: 240, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 240, H: 240, R: 1, Alpha: 1},
		},
		Order: []int{1, 0},
	}

	c, ok := displayBackground(display)
	if !ok {
		t.Fatal("no background")
	}

	r, _, _, _ := c.RGBA()
	if r>>8 != 255 {
		t.Fatalf("picked the wrong op: red = %d", r>>8)
	}
}

func TestPageBackgroundFallback(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{width: 100, height: 100}}
	if c := s.pageBackground(); c != color.White {
		t.Fatalf("fallback = %v", c)
	}
}

func TestBadgeRectSitsInTheTopRight(t *testing.T) {
	t.Parallel()

	const screenW = 640.0
	x, y, width, height := badgeRect(screenW, badgeLabel)

	if width <= 0 || height <= 0 {
		t.Fatalf("badge size = %v x %v, want positive", width, height)
	}

	if got := x + width + 8; got != screenW {
		t.Fatalf("badge right edge = %v, want %v", got, screenW)
	}

	if y != 8 {
		t.Fatalf("badge top = %v, want 8", y)
	}
}

func TestBitmapViewportReusesGPUBufferUntilResize(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body style="margin:0"><div style="height:400px;background:red;` +
		`border-radius:50% / 25%"></div></body>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	s := &shell{app: p, fallback: true, screenW: 100, screenH: 100}
	if err := s.prepareBitmapViewport(); err != nil {
		t.Fatal(err)
	}
	first := s.bitmapView.img
	if first == nil {
		t.Fatal("no viewport bitmap")
	}
	s.scrollY = 50
	if err := s.prepareBitmapViewport(); err != nil {
		t.Fatal(err)
	}
	if s.bitmapView.img != first {
		t.Fatal("scroll recreated GPU image")
	}
	s.screenW = 120
	p.SetSize(120, 100)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.prepareBitmapViewport(); err != nil {
		t.Fatal(err)
	}
	if s.bitmapView.img == first {
		t.Fatal("resize retained wrong buffer size")
	}
	s.bitmapView.img.Deallocate()
}

type fakeSlicerScreen struct {
	*fakeScreen
	sliceReads int
}

func (s *fakeSlicerScreen) BoxesSlice() []layout.Box {
	s.sliceReads++
	return s.boxes
}

var _ host.BoxSlicer = (*fakeSlicerScreen)(nil)

func TestShellBoxesPreference(t *testing.T) {
	displayBoxes := []layout.Box{{ID: "disp", X: 10, Y: 10, W: 50, H: 50}}
	slicerBoxes := []layout.Box{{ID: "slicer", X: 0, Y: 0, W: 100, H: 100}}
	fallbackBoxes := []layout.Box{{ID: "fallback", X: 0, Y: 0, W: 200, H: 200}}

	slicerApp := &fakeSlicerScreen{
		fakeScreen: &fakeScreen{boxes: slicerBoxes},
	}
	s := &shell{
		app: slicerApp,
		display: &layout.Display{
			Boxes: displayBoxes,
		},
	}

	// 1. When display has boxes, s.boxes() returns s.display.Boxes directly.
	got := s.boxes()
	if len(got) != 1 || got[0].ID != "disp" || &got[0] != &displayBoxes[0] {
		t.Fatalf("expected display boxes, got: %+v", got)
	}
	if slicerApp.sliceReads != 0 {
		t.Fatalf("sliceReads = %d, want 0 when display.Boxes is present", slicerApp.sliceReads)
	}

	// 2. When display is nil, s.boxes() calls host.BoxSlicer.
	s.display = nil
	got = s.boxes()
	if len(got) != 1 || got[0].ID != "slicer" || &got[0] != &slicerBoxes[0] {
		t.Fatalf("expected slicer boxes, got: %+v", got)
	}
	if slicerApp.sliceReads != 1 {
		t.Fatalf("sliceReads = %d, want 1", slicerApp.sliceReads)
	}

	// 3. Fallback when app doesn't implement host.BoxSlicer
	fallbackApp := &fakeScreen{boxes: fallbackBoxes}
	s.app = fallbackApp
	got = s.boxes()
	if len(got) != 1 || got[0].ID != "fallback" {
		t.Fatalf("expected fallback boxes, got: %+v", got)
	}

	// 4. boxAt hit-testing
	box, ok := s.boxAt(50, 50)
	if !ok || box.ID != "fallback" {
		t.Fatalf("boxAt(50, 50) = %+v, %v, want fallback", box, ok)
	}
	_, ok = s.boxAt(500, 500)
	if ok {
		t.Fatal("boxAt(500, 500) want false")
	}

	// 5. Zero allocations for s.boxes() and pointer-move hit-testing
	s.app = slicerApp
	s.display = &layout.Display{Boxes: displayBoxes}

	allocs := testing.AllocsPerRun(10, func() {
		_ = s.boxes()
		_, _ = s.boxAt(20, 20)
		s.devHover(20, 20)
	})
	if allocs != 0 {
		t.Fatalf("shell pointer-move operations allocated %f times, want 0", allocs)
	}
}

func BenchmarkPointerMoveBoxHit(b *testing.B) {
	boxes := make([]layout.Box, 100)
	for i := range boxes {
		boxes[i] = layout.Box{
			ID: "box",
			X:  float64(i * 10),
			Y:  float64(i * 10),
			W:  50,
			H:  50,
		}
	}
	s := &shell{
		display: &layout.Display{Boxes: boxes},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64((i % 100) * 10)
		s.devHover(x, x)
	}
}

func TestPinnedChromeDoesNotExtendScroll(t *testing.T) {
	screen, s := newPinShell()
	// IDs deliberately differ from operation indexes.
	s.display.Ops[0].ID = 19
	s.display.Ops[3].ID = 73
	s.display.Ops[3].Y = 900
	screen.boxes = []layout.Box{
		{Y: 0, H: 400, OperationIDs: []uint64{19}},
		{Y: 900, H: 40, OperationIDs: []uint64{73}},
	}
	_, h := s.contentSize()
	if h != 400 {
		t.Fatalf("pinned chrome extended document to %d, want 400", h)
	}
	s.scrollY = 700
	s.pullScroll()
	if s.scrollY != 200 {
		t.Fatalf("scroll offset %d, want 200", s.scrollY)
	}
}

// TestContentSizeIgnoresClippedOverflow pins that paint clipped by an
// overflow ancestor does not extend the scrollable document: a marquee
// track twice its lane must not summon a horizontal scrollbar.
func TestContentSizeIgnoresClippedOverflow(t *testing.T) {
	t.Parallel()

	fake := &fakeScreen{width: 1600, height: 900}
	s := newShell(context.Background(), fake)

	fake.boxes = []layout.Box{
		{ID: "page", Tag: "div", X: 0, Y: 0, W: 1600, H: 900},
		{
			ID: "tape-track", Tag: "div", X: 400, Y: 860, W: 2097, H: 28,
			OverflowClip:    [4]float64{400, 860, 800, 28},
			HasOverflowClip: true,
			OperationIDs:    []uint64{7},
			ParentIndex:     0,
		},
	}

	w, h := s.contentSize()
	if w != 1600 || h != 900 {
		t.Fatalf("content = %dx%d, want 1600x900 (clipped track ignored)", w, h)
	}

	unclipped := &fakeScreen{width: 1600, height: 900}
	s2 := newShell(context.Background(), unclipped)
	unclipped.boxes = []layout.Box{
		{ID: "page", Tag: "div", X: 0, Y: 0, W: 1600, H: 900},
		{ID: "wide", Tag: "div", X: 0, Y: 0, W: 2000, H: 900, OperationIDs: []uint64{9}},
	}

	w, _ = s2.contentSize()
	if w != 2000 {
		t.Fatalf("unclipped content width = %d, want 2000", w)
	}
}

func TestFramePointScalesOntoThePicture(t *testing.T) {
	t.Parallel()

	x, y := framePoint(50, 25, 200, 100, 100, 50)
	if x != 100 || y != 50 {
		t.Fatalf("point = %v, %v", x, y)
	}
}

func TestFramePointMatchesWhenSizesMatch(t *testing.T) {
	t.Parallel()

	x, y := framePoint(12, 34, 480, 640, 480, 640)
	if x != 12 || y != 34 {
		t.Fatalf("point = %v, %v", x, y)
	}
}

func TestPanScrollClampsToThePage(t *testing.T) {
	t.Parallel()

	x, y := panScroll(0, 0, 0, 1, 200, 800, 200, 400)
	if x != 0 || y != 0 {
		t.Fatalf("wheel up at the top = %d, %d", x, y)
	}

	x, y = panScroll(0, 0, 0, -1, 200, 800, 200, 400)
	if x != 0 || y != scrollStep {
		t.Fatalf("wheel down = %d, %d", x, y)
	}

	x, y = panScroll(0, 400, 0, -1, 200, 800, 200, 400)
	if x != 0 || y != 400 {
		t.Fatalf("past the end = %d, %d", x, y)
	}
}

func TestContentPointAddsTheScroll(t *testing.T) {
	t.Parallel()

	x, y := contentPoint(10, 20, 0, 100, false, 200, 800, 200, 400)
	if x != 10 || y != 120 {
		t.Fatalf("scrolled point = %v, %v", x, y)
	}
}

func TestFramePointRejectsAnEmptyScreen(t *testing.T) {
	t.Parallel()

	x, y := framePoint(12, 34, 480, 640, 0, 640)
	if x != 0 || y != 0 {
		t.Fatalf("point = %v, %v", x, y)
	}
}

func TestNormalizeMaxFPS(t *testing.T) {
	if got := NormalizeMaxFPS(0); got != 0 {
		t.Fatalf("zero = %d, want uncapped 0", got)
	}
	if got := NormalizeMaxFPS(-1); got != 0 {
		t.Fatalf("uncapped = %d, want 0", got)
	}
	for _, fps := range []int{60, 120, 144, 240} {
		if got := NormalizeMaxFPS(fps); got != fps {
			t.Fatalf("fps %d = %d", fps, got)
		}
	}
	if got := NormalizeMaxFPS(10); got != 0 {
		t.Fatalf("invalid low = %d, want uncapped", got)
	}
	if got := NormalizeMaxFPS(29); got != 0 {
		t.Fatalf("below min = %d, want uncapped", got)
	}
	for _, fps := range []int{30, 31, 999, 1000} {
		if got := NormalizeMaxFPS(fps); got != fps {
			t.Fatalf("boundary fps %d = %d", fps, got)
		}
	}
	if got := NormalizeMaxFPS(1001); got != 0 {
		t.Fatalf("above max = %d, want uncapped", got)
	}
	if got := NormalizeMaxFPS(5000); got != 0 {
		t.Fatalf("invalid high = %d, want uncapped", got)
	}
}

func TestLoopRunsHotterThanPresents(t *testing.T) {
	for _, fps := range []int{60, 120, 144, 240} {
		if got := LoopFPS(fps); got != fps+FPSHeadroom {
			t.Fatalf("loop(%d) = %d, want %d", fps, got, fps+FPSHeadroom)
		}
		if pi, li := PresentInterval(fps), LoopInterval(fps); pi != li {
			t.Fatalf("fps %d: present %v must equal loop %v (loop paced to cap)", fps, pi, li)
		}
	}
	if LoopFPS(0) != 0 || PresentInterval(0) != 0 || LoopInterval(0) != 0 {
		t.Fatal("uncapped must have zero intervals")
	}
}

func TestCapHoldsUnderSustainedLoad(t *testing.T) {
	const maxFPS = 60
	base := time.Unix(1000, 0)
	lastLoop := base
	presents := 0
	at := base
	// Simulate 2s of sustained load paced to the cap: 120 loop iterations
	// at 16.6ms, each presenting once.
	for i := 0; i < 120; i++ {
		at = at.Add(LoopInterval(maxFPS))
		lastLoop = at
		presents++
	}
	if presents != 120 {
		t.Fatalf("2s at 60 cap presented %d, want 120", presents)
	}
	_ = lastLoop
	// The hotter app rate is carried by TPS headroom, not extra presents:
	// 60 presents, 80 updates per second.
	if LoopFPS(maxFPS) != maxFPS+FPSHeadroom {
		t.Fatalf("loop TPS = %d, want %d", LoopFPS(maxFPS), maxFPS+FPSHeadroom)
	}
}

func TestLoopWaitEnforcesIntervalWithoutBusySpin(t *testing.T) {
	const maxFPS = 60
	base := time.Unix(2000, 0)
	if got := loopWait(maxFPS, time.Time{}, base); got != 0 {
		t.Fatalf("first frame wait = %v, want 0", got)
	}
	if got := loopWait(maxFPS, base, base); got != LoopInterval(maxFPS) {
		t.Fatalf("immediate wait = %v, want %v", got, LoopInterval(maxFPS))
	}
	if got := loopWait(maxFPS, base, base.Add(2*LoopInterval(maxFPS))); got != 0 {
		t.Fatalf("late wait = %v, want 0", got)
	}
	if got := loopWait(0, base, base); got != 0 {
		t.Fatalf("uncapped wait = %v, want 0", got)
	}
}

func TestCoalescingNeverStarvesImmediateRequests(t *testing.T) {
	// The Draw limiter paces the loop to the present cap with sleep; it
	// never skips heavy work, so an animation asking for the next frame at
	// once (WantFrameIn(0)) is always presented on the next loop tick. This
	// pins the contract: every loop iteration presents, and the hotter app
	// rate is carried by TPS headroom, not skipped presents.
	const maxFPS = 60
	if LoopInterval(maxFPS) != PresentInterval(maxFPS) {
		t.Fatal("loop must be paced to the present cap")
	}
}

// optInScreen is a ticking screen whose frame-dirty flag clears on read, the
// way *page.Page does.
type optInScreen struct {
	*fakeScreen
	ticking bool
	dirty   bool
	reads   int
}

func (o *optInScreen) Ticking() bool { return o.ticking }

func (o *optInScreen) FrameDirty() bool {
	o.reads++
	on := o.dirty
	o.dirty = false

	return on
}

// TestFullReplayCapturesOptInPerTick pins the window-side fix: the opt-in is
// read once per Update, so a Draw that Ebiten issues without an intervening
// Update keeps the buffer path instead of replaying the whole list.
func TestFullReplayCapturesOptInPerTick(t *testing.T) {
	app := &optInScreen{fakeScreen: &fakeScreen{width: 100, height: 100}, ticking: true, dirty: true}
	s := &shell{app: app}

	s.captureFrameOptIn()
	if s.fullReplay() {
		t.Fatal("an opted-in tick must keep the buffer path")
	}

	reads := app.reads
	if s.fullReplay() {
		t.Fatal("a draw without an Update must not fall back to the full replay")
	}
	if app.reads != reads {
		t.Fatalf("FrameDirty read %d times during Draw, want %d", app.reads, reads)
	}

	app.dirty = false
	s.captureFrameOptIn()
	if s.fullReplay() {
		t.Fatal("a Page ticker without the opt-in stays on the buffer path")
	}
}

func TestFullReplayIgnoresNonTickingScreens(t *testing.T) {
	app := &optInScreen{fakeScreen: &fakeScreen{width: 10, height: 10}}
	s := &shell{app: app}
	s.captureFrameOptIn()

	if s.fullReplay() {
		t.Fatal("a screen without a tick keeps the ordinary partial path")
	}
}

func (g *partialFrameGame) check(bounds image.Rectangle, pixels []byte) {
	if g.err != nil {
		return
	}

	green := [4]byte{0, 255, 0, 255}
	red := [4]byte{255, 0, 0, 255}
	white := [4]byte{255, 255, 255, 255}
	corner := bounds.Max.Sub(image.Pt(10, 10))
	outside := g.at(bounds, pixels, corner.X, corner.Y)
	inside := bounds.Min.Add(image.Pt(bounds.Dx()/5, bounds.Dy()/5))
	scale := bounds.Dx() / 100

	if g.frame == 0 {
		file, err := os.Open("testdata/partial-frame.png")
		if err != nil {
			g.err = err
			return
		}
		golden, err := png.Decode(file)
		closeErr := file.Close()
		if err != nil {
			g.err = err
			return
		}
		if closeErr != nil {
			g.err = closeErr
			return
		}
		if bounds.Dx() != 100 || bounds.Dy() != 100 {
			g.err = fmt.Errorf("golden size differs: %v", bounds)
			return
		}
		for y := range 100 {
			for x := range 100 {
				r, gv, b, a := golden.At(x, y).RGBA()
				got := g.at(bounds, pixels, x, y)
				const channelScale = 257
				differs := r != uint32(got[0])*channelScale || gv != uint32(got[1])*channelScale ||
					b != uint32(got[2])*channelScale || a != uint32(got[3])*channelScale
				if differs {
					g.err = fmt.Errorf("golden differs at %d,%d: %v != %v", x, y, got, golden.At(x, y))
					return
				}
			}
		}
	}
	switch g.frame {
	case 0, 3:
		if outside == green {
			g.err = fmt.Errorf("full frame %d left the sentinel painted", g.frame)
		}
	case 1, 2:
		if outside != green {
			g.err = fmt.Errorf("partial frame filled the screen: pixel = %v", outside)
		}

		if g.at(bounds, pixels, inside.X, inside.Y) != red {
			g.err = fmt.Errorf("partial frame did not blit the red ink")
		}
		for y := 10 * scale; y < 30*scale; y++ {
			for x := 10 * scale; x < 30*scale; x++ {
				want := white
				if x >= 12*scale && y >= 12*scale {
					want = red
				}
				if got := g.at(bounds, pixels, x, y); got != want {
					g.err = fmt.Errorf("partial damage differs at %d,%d: %v != %v", x, y, got, want)
					return
				}
			}
		}
	case 4:
		// The ink sits at 12..32 px before the scroll. After a 20 px scroll
		// up, row 5 shows the box and row 35 shows the page under it.
		if got := g.at(bounds, pixels, 15*scale, 5*scale); got != red {
			g.err = fmt.Errorf("scroll copy lost the scrolled ink: pixel = %v", got)
		}

		if got := g.at(bounds, pixels, 15*scale, 35*scale); got != white {
			g.err = fmt.Errorf("scroll copy left the old ink: pixel = %v", got)
		}
	default:
		g.checkDock(bounds, pixels, scale)
	}
}

func (g *partialFrameGame) at(bounds image.Rectangle, pixels []byte, x, y int) [4]byte {
	i := (y*bounds.Dx() + x) * 4

	return [4]byte{pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]}
}

// checkDock guards frames 5 to 7: the dock hides the box under it, and
// closing it shows the box's current colour. The box's points scale to
// 80..100 x 40..60 pixels, so 90,50 is inside it.
func (g *partialFrameGame) checkDock(bounds image.Rectangle, pixels []byte, scale int) {
	if g.err != nil {
		return
	}

	red := [4]byte{255, 0, 0, 255}
	blue := [4]byte{0, 0, 255, 255}

	switch g.frame {
	case 5:
		if got := g.at(bounds, pixels, 15*scale, 15*scale); got != red {
			g.err = fmt.Errorf("dock frame 5 lost the page box: pixel = %v", got)
		}

		if got := g.at(bounds, pixels, 90*scale, 50*scale); got == red {
			g.err = fmt.Errorf("dock frame 5 left the hidden box showing: pixel = %v", got)
		}
	case 6:
		if got := g.at(bounds, pixels, 90*scale, 50*scale); got == blue {
			g.err = fmt.Errorf("dock frame 6 showed the hidden box: pixel = %v", got)
		}
	case 7:
		if got := g.at(bounds, pixels, 90*scale, 50*scale); got != blue {
			g.err = fmt.Errorf("dock frame 7 kept the stale strip: pixel = %v", got)
		}
	}
}

func (g *partialFrameGame) prepare() {
	switch g.frame {
	case 0, 1:
		g.app.click(image.Rect(10, 10, 30, 30))
	case 2:
		if g.shell.partial.buf != nil {
			g.shell.partial.buf.Deallocate()
			g.shell.partial.buf = nil
		}
	case 3:
		g.app.ok = false
		g.app.rects = nil
		g.shell.scrollY = 20
	case 4:
		g.shell.scrollY = 0
		g.shell.dev.on = true
		g.shell.dev.dockW = 40
		g.setDockPage(1, 0, 0)
	case 5:
		g.setDockPage(0, 0, 1)
	case 6:
		g.shell.dev.on = false
		g.app.rects, g.app.ok = nil, false
	}
}

// setDockPage swaps in the page whose box sits under the dock and reports
// its damage.
func (g *partialFrameGame) setDockPage(r, gv, b float64) {
	d := dockDisplay(r, gv, b)
	g.app.display, g.shell.display = d, d
	g.app.click(image.Rect(60, 30, 75, 45))
}

// dockDisplay is the 100 x 100 page plus one box under the dock's left edge.
// The box's points scale to 80..100 x 40..60 pixels, right of the 40 px dock.
func dockDisplay(r, g, b float64) *layout.Display {
	d := partialDisplay(100, 100)
	d.Ops = append(d.Ops, layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 60, Y: 30, W: 15, H: 15,
		R: r, G: g, B: b, Alpha: 1,
	})
	d.Order = []int{0, 1, 2}

	return d
}

// TestPartialFrameLeavesTheScreen is the pixel guard for the swap: a partial
// frame blits only its planned rects, so a sentinel painted outside them
// survives, a full frame fills the whole screen and wipes it, a scroll frame
// copies the whole buffer at the new offset, and a dock frame hides the page
// under the panel without repainting it until the dock closes.
func TestPartialFrameLeavesTheScreen(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("requires a graphics session; set OWNFRAME_FRAME_GPU_TEST=1")
	}

	app := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	app.display = partialDisplay(100, 100)
	app.gen = 1

	shell := newDirtyShell(app)
	shell.ctx = context.Background()
	game := &partialFrameGame{shell: shell, app: app}

	ebiten.SetWindowSize(100, 100)
	ebiten.SetScreenClearedEveryFrame(false)

	if err := ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}

	if game.err != nil {
		t.Fatal(game.err)
	}
}

// partialFrameGame paints a sentinel under the shell, then checks the pixels
// the frame left. Frame 0 is the first frame, which is full. Frames 1 and 2
// report damage, so they are partial. Frame 3 drops the buffer, which forces
// another full frame. Frame 4 scrolls without damage, so the frame copies the
// whole buffer. Frames 5 to 7 open a 40 px dock, damage the box under it, and
// close the dock. Each Draw prepares the next frame, because Ebiten may skip
// an Update between draws.
type partialFrameGame struct {
	shell *shell
	app   *dirtyScreen
	frame int
	err   error
}

func (g *partialFrameGame) Layout(int, int) (int, int) { return 100, 100 }

func (g *partialFrameGame) Update() error {
	// The focus-regain guard would turn a focus change into a full frame.
	// Keep the recorded focus current so the damage frames stay partial.
	g.shell.partial.focus = ebiten.IsFocused()

	if g.frame >= 8 {
		return ebiten.Termination
	}

	return nil
}

func (g *partialFrameGame) Draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{G: 255, A: 255})
	g.shell.Draw(dst)

	bounds := dst.Bounds()
	pixels := make([]byte, 4*bounds.Dx()*bounds.Dy())
	dst.ReadPixels(pixels)
	g.check(bounds, pixels)
	g.prepare()
	g.frame++
}

// stripes is n disjoint 5x5 rects in a row, for the cap cases.
func stripes(n int) []image.Rectangle {
	rects := make([]image.Rectangle, n)
	for i := range rects {
		rects[i] = image.Rect(i*10, 0, i*10+5, 5)
	}

	return rects
}

// planRects turns rects into repaintRect plans in order.
func planRects(rects []image.Rectangle) []repaintPlan {
	plans := make([]repaintPlan, len(rects))
	for i, r := range rects {
		plans[i] = repaintPlan{mode: repaintRect, rect: r}
	}

	return plans
}

func TestBlitPlanDecidesTheFrame(t *testing.T) {
	t.Parallel()

	full := image.Rect(0, 0, 100, 100)
	one := image.Rect(10, 10, 30, 30)
	two := image.Rect(60, 60, 80, 80)
	near := image.Rect(20, 20, 40, 40)
	wide := image.Rect(35, 35, 55, 55)
	chrome := image.Rect(0, 0, 8, 100)
	blit := []repaintPlan{{mode: repaintBlit, rect: full}}
	buffer := []repaintPlan{{mode: repaintBuffer, rect: full}}

	cases := []struct {
		name   string
		plans  []repaintPlan
		chrome []image.Rectangle
		full   bool
		rects  []image.Rectangle
	}{
		{"no plans", nil, nil, false, nil},
		{"blit only", blit, nil, false, nil},
		{"blit keeps chrome", blit, []image.Rectangle{chrome}, false,
			[]image.Rectangle{chrome}},
		{"one rect", planRects([]image.Rectangle{one}), nil, false,
			[]image.Rectangle{one}},
		{"two rects", planRects([]image.Rectangle{one, two}), nil, false,
			[]image.Rectangle{one, two}},
		{"overlapping rects", planRects([]image.Rectangle{one, near}), nil, false,
			[]image.Rectangle{one.Union(near)}},
		{"chrome unions a rect", planRects([]image.Rectangle{one}), []image.Rectangle{near}, false,
			[]image.Rectangle{one.Union(near)}},
		{"overlap through a third", planRects([]image.Rectangle{one, near, wide}), nil, false,
			[]image.Rectangle{one.Union(near).Union(wide)}},
		{"empty rects drop", planRects([]image.Rectangle{image.Rectangle{}, one}),
			[]image.Rectangle{image.Rectangle{}}, false, []image.Rectangle{one}},
		{"buffer plan", buffer, []image.Rectangle{chrome}, true, nil},
		{"at the cap", planRects(stripes(8)), nil, false, stripes(8)},
		{"past the cap", planRects(stripes(9)), nil, false,
			[]image.Rectangle{image.Rect(0, 0, 85, 5)}},
	}

	for _, c := range cases {
		full, rects := blitPlan(c.plans, c.chrome)
		if full != c.full {
			t.Fatalf("%s: full = %v, want %v", c.name, full, c.full)
		}

		if !slices.Equal(rects, c.rects) {
			t.Fatalf("%s: rects = %v, want %v", c.name, rects, c.rects)
		}
	}
}

// frameRects copies the committed chrome set so later frames cannot change
// the value under a check.
func frameRects(s *shell) []image.Rectangle {
	return slices.Clone(s.frameRestore())
}

// TestFrameBeginClearsCurrent drops a note made before begin: only the rects
// collected inside the open frame reach the commit.
func TestFrameBeginClearsCurrent(t *testing.T) {
	t.Parallel()

	s := &shell{}
	s.frameNote(image.Rect(1, 1, 2, 2))
	s.frameBegin()
	s.frameNote(image.Rect(3, 3, 4, 4))
	s.frameCommit()

	want := []image.Rectangle{image.Rect(3, 3, 4, 4)}
	if got := frameRects(s); !slices.Equal(got, want) {
		t.Fatalf("committed = %v, want %v", got, want)
	}
}

// TestFrameNoteRecordsOnlyPaintedRects keeps the non-empty notes in order
// and drops every empty one.
func TestFrameNoteRecordsOnlyPaintedRects(t *testing.T) {
	t.Parallel()

	s := &shell{}
	s.frameBegin()
	s.frameNote(image.Rect(10, 10, 30, 30))
	s.frameNote(image.Rect(5, 5, 5, 40))
	s.frameNote(image.Rectangle{})
	s.frameNote(image.Rectangle{Min: image.Pt(40, 40), Max: image.Pt(30, 30)})
	s.frameNote(image.Rect(50, 50, 70, 70))
	s.frameCommit()

	want := []image.Rectangle{image.Rect(10, 10, 30, 30), image.Rect(50, 50, 70, 70)}
	if got := frameRects(s); !slices.Equal(got, want) {
		t.Fatalf("committed = %v, want %v", got, want)
	}
}

// TestFrameRestoreKeepsTheCommittedFrame reads restore while the next frame
// is being collected: new notes must not leak into the committed set.
func TestFrameRestoreKeepsTheCommittedFrame(t *testing.T) {
	t.Parallel()

	s := &shell{}
	old := image.Rect(1, 1, 2, 2)
	s.frameBegin()
	s.frameNote(old)
	s.frameCommit()

	s.frameBegin()
	s.frameNote(image.Rect(3, 3, 4, 4))
	s.frameNote(image.Rect(5, 5, 6, 6))

	want := []image.Rectangle{old}
	if got := frameRects(s); !slices.Equal(got, want) {
		t.Fatalf("restore during collection = %v, want %v", got, want)
	}
}

// TestFrameCommitSwapsTheSets checks seen and that each commit replaces the
// restore set with the rects of this frame alone.
func TestFrameCommitSwapsTheSets(t *testing.T) {
	t.Parallel()

	s := &shell{}
	if s.frameSeen() {
		t.Fatal("frameSeen before the first commit")
	}

	first := image.Rect(1, 1, 2, 2)
	s.frameBegin()
	s.frameNote(first)
	s.frameCommit()

	if !s.frameSeen() {
		t.Fatal("frameCommit left seen false")
	}

	if got := frameRects(s); !slices.Equal(got, []image.Rectangle{first}) {
		t.Fatalf("first commit = %v, want %v", got, first)
	}

	second := image.Rect(3, 3, 4, 4)
	s.frameBegin()
	s.frameNote(second)
	s.frameCommit()

	if got := frameRects(s); !slices.Equal(got, []image.Rectangle{second}) {
		t.Fatalf("second commit = %v, want only %v", got, second)
	}
}

// TestFrameSteadyStateAllocatesNothing pins the slice swap: a frame that
// notes up to two rects reuses both backing arrays.
func TestFrameSteadyStateAllocatesNothing(t *testing.T) {
	s := &shell{}
	rects := []image.Rectangle{image.Rect(1, 1, 2, 2), image.Rect(3, 3, 4, 4)}
	frame := func() {
		s.frameBegin()
		for _, r := range rects {
			s.frameNote(r)
		}
		s.frameCommit()
	}
	frame()

	if n := testing.AllocsPerRun(100, frame); n != 0 {
		t.Fatalf("allocs per frame = %v, want 0", n)
	}
}

// replay.LayerSource and host.LayerSource must stay one type. The shell asserts
// the replay role on the screen it holds, and that screen is *page.Page.
var _ replay.LayerSource = (*page.Page)(nil)

// windowPageLayerHTML animates one element's transform, the entry the page
// promotes today.
const windowPageLayerHTML = `<style>
@keyframes slide { 0% { transform: translateX(0px); } 100% { transform: translateX(120px); } }
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	animation-name: slide; animation-duration: 2s;
	animation-timing-function: linear; animation-iteration-count: infinite;
	animation-fill-mode: both; }
</style><div id="target"></div>`

// TestWindowFindsPageLayerSource proves the window reaches the page through its
// LayerSource assertion after the type move, not through a wrapper.
func TestWindowFindsPageLayerSource(t *testing.T) {
	t.Parallel()

	p, err := page.New(page.Config{HTML: windowPageLayerHTML, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}

	shell := &shell{app: p}
	if _, ok := shell.app.(replay.LayerSource); !ok {
		t.Fatal("the shell's screen does not satisfy replay.LayerSource")
	}
	specs := shell.layerSpecs(p.Display())
	if len(specs) != 1 {
		t.Fatalf("window layer specs = %d, want one", len(specs))
	}
	if specs[0].ElementID != "target" {
		t.Fatalf("element = %q, want target", specs[0].ElementID)
	}
	if specs[0].Matrix.IsIdentity() {
		t.Fatal("the window received an identity layer matrix")
	}
}

const repaintWidth, repaintHeight, repaintStageCount = 600, 300, 8

// TestWindowRepaintPixels compares the retained window framebuffer against a
// fresh full replay. Both painting and readback run inside Ebiten's Draw loop.
func TestWindowRepaintPixels(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check for an isolated graphics session")
	}
	ctx := context.Background()
	p, err := page.New(page.Config{
		Width: repaintWidth, Height: repaintHeight,
		HTML: `<body style="margin:0;background:white"><div style="height:900px">
<div style="position:absolute;left:240px;top:70px;width:80px;height:80px;background:{{.Color}};border-radius:8px"></div>
{{if .Visible}}<div style="position:absolute;left:40px;top:160px;width:60px;height:60px;background:blue"></div>{{end}}
<div style="position:fixed;z-index:10;left:0;top:0;width:600px;height:24px;background:black"></div>
</div></body>`,
	})
	if err != nil {
		t.Fatal(err)
	}
	view := repaintView{Color: "red", Visible: true}
	p.SetData(view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	g := &repaintPixelGame{page: p, shell: newShell(ctx, p), view: view}
	defer g.shell.cancel()
	ebiten.SetWindowSize(repaintWidth, repaintHeight)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if g.step != repaintStageCount {
		t.Fatalf("only completed %d of %d stages", g.step, repaintStageCount)
	}
}

func TestWindowBackdropFilterReplaysFullFrame(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check for an isolated graphics session")
	}

	ctx := context.Background()
	const width, height = 180, 120
	p, err := page.New(page.Config{
		Width: width, Height: height,
		HTML: `<body style="margin:0"><div style="position:absolute;left:0;top:0;width:180px;height:120px;background:{{.Color}}"></div>` +
			`<div style="position:absolute;left:28px;top:20px;width:124px;height:80px;z-index:1;` +
			`backdrop-filter:blur(12px) saturate(1.25);background:rgba(255,255,255,.2)"></div></body>`,
	})
	if err != nil {
		t.Fatal(err)
	}
	view := struct{ Color string }{Color: "#e03020"}
	p.SetData(view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	g := &backdropReplayGame{page: p, shell: newShell(ctx, p), view: view}
	defer g.shell.cancel()
	ebiten.SetWindowSize(width, height)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if g.step != 2 {
		t.Fatalf("completed %d frames, want 2", g.step)
	}
}

type backdropReplayGame struct {
	page  *page.Page
	shell *shell
	view  struct{ Color string }
	step  int
	err   error
}

func (g *backdropReplayGame) Layout(int, int) (int, int) { return 180, 120 }
func (g *backdropReplayGame) Update() error {
	if g.err != nil || g.step == 2 {
		return ebiten.Termination
	}
	return nil
}

func (g *backdropReplayGame) Draw(dst *ebiten.Image) {
	if g.err != nil || g.step == 2 {
		return
	}
	if err := g.shell.syncImage(); err != nil {
		g.err = err
		return
	}
	g.shell.Draw(dst)
	if g.shell.partial.tiles != nil {
		g.err = fmt.Errorf("filtered frame %d allocated a tile grid", g.step)
		return
	}
	if g.shell.partial.mode != repaintBlit || g.shell.partial.whole || !g.shell.partial.live {
		g.err = fmt.Errorf("filtered frame %d state = mode %v whole %t live %t", g.step,
			g.shell.partial.mode, g.shell.partial.whole, g.shell.partial.live)
		return
	}

	expected := ebiten.NewImage(180, 120)
	defer expected.Deallocate()
	g.shell.paintFullReplay(expected, g.shell.display)
	actualPixels, expectedPixels := make([]byte, 180*120*4), make([]byte, 180*120*4)
	dst.ReadPixels(actualPixels)
	expected.ReadPixels(expectedPixels)
	if !bytes.Equal(actualPixels, expectedPixels) {
		g.err = fmt.Errorf("filtered frame %d differs from full replay", g.step)
		return
	}

	if g.step == 0 {
		g.view.Color = "#2050d0"
		g.page.SetData(g.view)
		if err := g.page.Redraw(context.Background()); err != nil {
			g.err = err
			return
		}
	}
	g.step++
}

type repaintView struct {
	Color   string
	Visible bool
}

// Repeating a clean frame catches accidental clearing of untouched pixels.
func repaintStageName(i int) string {
	return [...]string{"initial", "color across tiles", "remove element", "scroll with fixed header", "clean frame", "change while scrolled", "cleared mobile surface", "clean mobile surface"}[i]
}

type repaintPixelGame struct {
	page     *page.Page
	shell    *shell
	view     repaintView
	expected *ebiten.Image
	step     int
	err      error
}

func (g *repaintPixelGame) Layout(int, int) (int, int) { return repaintWidth, repaintHeight }
func (g *repaintPixelGame) Update() error {
	if g.err != nil || g.step == repaintStageCount {
		return ebiten.Termination
	}
	return nil
}

func (g *repaintPixelGame) Draw(dst *ebiten.Image) {
	if g.err != nil || g.step == repaintStageCount {
		return
	}
	if err := g.shell.syncImage(); err != nil {
		g.err = err
		return
	}
	if g.step >= 6 {
		dst.Clear()
	}
	g.shell.partial.focus = ebiten.IsFocused()
	g.shell.Draw(dst)
	if g.shell.partial.tiles == nil {
		w, h := g.shell.contentSize()
		g.err = fmt.Errorf("%s bypassed tiles: content=%dx%d screen=%dx%d fallback=%v stretched=%v display=%v safe=%v", repaintStageName(g.step), w, h, g.shell.screenW, g.shell.screenH, g.shell.fallback, g.shell.stretched(), g.shell.display != nil, g.shell.bufferOrderSafe())
		return
	}
	if g.step == 1 && g.shell.partial.mode != repaintRect {
		g.err = fmt.Errorf("color change did not take partial repaint path")
		return
	}
	if g.expected == nil {
		g.expected = ebiten.NewImage(repaintWidth, repaintHeight)
	}
	g.shell.paintFullReplay(g.expected, g.shell.display)
	g.shell.drawScrollbars(g.expected)
	actual, expected := make([]byte, repaintWidth*repaintHeight*4), make([]byte, repaintWidth*repaintHeight*4)
	dst.ReadPixels(actual)
	g.expected.ReadPixels(expected)
	if !bytes.Equal(actual, expected) {
		for i := range actual {
			if actual[i] != expected[i] {
				pixel := i / 4
				x, y := pixel%repaintWidth, pixel/repaintWidth
				g.err = fmt.Errorf("%s: retained window differs at (%d,%d), channel %d: got %d want %d",
					repaintStageName(g.step), x, y, i%4, actual[i], expected[i])
				return
			}
		}
		g.err = fmt.Errorf("%s: retained window differs from full replay", repaintStageName(g.step))
		return
	}
	g.prepareNext()
	g.step++
}

func (g *repaintPixelGame) prepareNext() {
	const removeStage, scrollStage, cleanStage = 1, 2, 4
	switch g.step {
	case 0:
		g.view.Color = "green"
	case removeStage:
		g.view.Visible = false
	case scrollStage:
		const offset = 90
		g.shell.scrollY = offset
		g.page.SetScrollOffset(0, offset)
		return
	case cleanStage:
		g.view.Color = "purple"
	case 5:
		ebiten.SetScreenClearedEveryFrame(true)
		return
	default:
		return
	}
	g.page.SetData(g.view)
	g.err = g.page.Redraw(context.Background())
}

func TestStickyViewportPixelsAndHits(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check")
	}
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	fixture, err := os.ReadFile("testdata/sticky.html")
	if err != nil {
		t.Fatal(err)
	}
	p, err := page.New(page.Config{Width: 200, Height: 100, Scrollbar: host.ScrollbarHidden, HTML: string(fixture)})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData("red")
	clicked := ""
	p.Handle(page.Handlers{Click: func(_ context.Context, box page.Box) error { clicked = box.Action; return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	shell := newShell(ctx, p)
	defer shell.cancel()
	defer shell.closeA11y()
	game := &stickyPixelGame{shell: shell, page: p, ctx: ctx, test: t, clicked: &clicked}
	ebiten.SetWindowSize(200, 100)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}
	if game.err != nil {
		t.Fatal(game.err)
	}
	if game.step != 10 {
		t.Fatal("sticky native scenario did not finish")
	}
}

type stickyPixelGame struct {
	shell    *shell
	page     *page.Page
	ctx      context.Context
	test     *testing.T
	clicked  *string
	step     int
	prepared bool
	err      error
}

func (g *stickyPixelGame) Layout(int, int) (int, int) { return 200, 100 }
func (g *stickyPixelGame) Update() error {
	if g.err != nil || g.step == 10 {
		return ebiten.Termination
	}
	if g.prepared {
		return nil
	}
	offsets := []int{0, 100, 200, 360, 400, 200, 200, 300, 300, 0}
	g.shell.scrollY = offsets[g.step]
	if g.step == 6 {
		g.page.SetData("yellow")
		if err := g.page.Redraw(g.ctx); err != nil {
			return err
		}
	}
	g.shell.pageZoom = 1
	if g.step == 7 || g.step == 8 {
		g.shell.pageZoom = 1.5
	}
	if err := g.shell.syncImage(); err != nil {
		return err
	}
	if err := g.shell.syncScrollWindow(); err != nil {
		return err
	}
	if g.step == 2 || g.step == 8 {
		x, y := g.shell.contentAt(10, 10, 200, 100)
		if err := g.page.Click(g.ctx, x, y); err != nil {
			return err
		}
		if *g.clicked != "head" {
			return fmt.Errorf("sticky screen point hit %q at zoom %g", *g.clicked, g.shell.zoom())
		}
		if g.step == 2 {
			x, y = g.shell.contentAt(175, 10, 200, 100)
			if err := g.page.Click(g.ctx, x, y); err != nil {
				return err
			}
			if *g.clicked != "cover" {
				return fmt.Errorf("higher-z overlap hit %q", *g.clicked)
			}
		}
	}
	g.prepared = true
	return nil
}
func (g *stickyPixelGame) Draw(screen *ebiten.Image) {
	if !g.prepared || g.err != nil {
		return
	}
	g.shell.Draw(screen)
	pixels := make([]byte, 200*100*4)
	screen.ReadPixels(pixels)
	sample := func(x, y int) [3]byte { at := (y*200 + x) * 4; return [3]byte{pixels[at], pixels[at+1], pixels[at+2]} }
	red, blue, yellow := [3]byte{255, 0, 0}, [3]byte{0, 0, 255}, [3]byte{255, 255, 0}
	pointY := 10
	want := red
	switch g.step {
	case 0:
		pointY = 90
	case 3:
		pointY = 10
	case 4:
		want = blue
	case 6, 7, 8:
		want = yellow
	case 9:
		pointY = 90
		want = yellow
	}
	if got := sample(10, pointY); got != want {
		g.err = fmt.Errorf("sticky step %d pixel %v want %v", g.step, got, want)
		return
	}
	if g.step == 2 && sample(175, 10) != [3]byte{0, 255, 0} {
		g.err = fmt.Errorf("higher-z overlap not painted above sticky")
		return
	}
	if g.step == 3 && sample(10, 30) != blue {
		g.err = fmt.Errorf("sticky containing-block release left old paint")
		return
	}
	if (g.step == 7 || g.step == 8) && sample(10, 65) != blue {
		g.err = fmt.Errorf("zoomed sticky paint has wrong height")
		return
	}
	if dir := os.Getenv("OWNFRAME_STICKY_CAPTURE_DIR"); dir != "" {
		//nolint:gosec // reason: the test runner chooses this local capture directory.
		if err := os.MkdirAll(dir, 0700); err != nil {
			g.err = err
			return
		}
		//nolint:gosec // reason: the test runner chooses this local capture directory.
		file, err := os.Create(filepath.Join(dir, fmt.Sprintf("sticky-%d.png", g.step)))
		if err != nil {
			g.err = err
			return
		}
		img := &image.RGBA{Pix: pixels, Stride: 200 * 4, Rect: image.Rect(0, 0, 200, 100)}
		err = png.Encode(file, img)
		closeErr := file.Close()
		if err != nil {
			g.err = err
			return
		}
		if closeErr != nil {
			g.err = closeErr
			return
		}
	}
	g.test.Logf("sticky step %d: scroll=%d zoom=%g pixel=%v click=%s", g.step, g.shell.scrollY, g.shell.zoom(), sample(10, pointY), *g.clicked)
	g.step++
	g.prepared = false
}

// tickScreen is a fakeScreen with a per-frame callback.
type tickScreen struct {
	*fakeScreen
	ticks int
	err   error
}

func (s *tickScreen) Tick(context.Context) error {
	s.ticks++

	return s.err
}

func TestTickFrameRunsTicker(t *testing.T) {
	t.Parallel()

	screen := &tickScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: screen, ctx: context.Background()}

	if err := s.tickFrame(); err != nil {
		t.Fatalf("tickFrame: %v", err)
	}

	if screen.ticks != 1 {
		t.Fatalf("ticks = %d, want 1", screen.ticks)
	}
}

func TestTickFrameSkipsPlainScreen(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{}, ctx: context.Background()}
	if err := s.tickFrame(); err != nil {
		t.Fatalf("tickFrame: %v", err)
	}
}

func TestTickFrameReturnsError(t *testing.T) {
	t.Parallel()

	want := errors.New("frame failed")
	screen := &tickScreen{fakeScreen: &fakeScreen{}, err: want}
	s := &shell{app: screen, ctx: context.Background()}

	if err := s.tickFrame(); !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
}

func TestTransparentWindowDoesNotAddBackground(t *testing.T) {
	s := &shell{transparent: true, display: &layout.Display{
		Width: 240, PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{{Kind: layout.DisplayOpFillRect, W: 180, H: 165, Alpha: 1}},
	}}
	if _, _, _, alpha := s.pageBackground().RGBA(); alpha != 0 {
		t.Fatalf("transparent window background alpha = %d", alpha)
	}

	s.transparent = false
	if _, _, _, alpha := s.pageBackground().RGBA(); alpha != 65535 {
		t.Fatalf("normal window background alpha = %d", alpha)
	}
}

func TestFitBoxShowsTheWholePage(t *testing.T) {
	scale, ox, oy := fitBox(200, 400, 100, 100)
	if scale != 0.25 || ox != 25 || oy != 0 {
		t.Fatalf("scale=%v ox=%v oy=%v", scale, ox, oy)
	}
}

func TestFitBoxLeavesAMatchingPage(t *testing.T) {
	scale, ox, oy := fitBox(1080, 2400, 1080, 2400)
	if scale != 1 || ox != 0 || oy != 0 {
		t.Fatalf("scale=%v ox=%v oy=%v", scale, ox, oy)
	}
}

type fittedScreen struct{ *requestScreen }

func (*fittedScreen) ViewLocked() bool { return true }

func TestLockedViewMapsFittedControlsAndRejectsScrolling(t *testing.T) {
	app := &fittedScreen{&requestScreen{fakeScreen: &fakeScreen{
		width: 100, height: 100,
		boxes: []layout.Box{{W: 200, H: 400}},
	}}}
	s := &shell{app: app, screenW: 100, screenH: 100, scrollY: 50}
	x, y := s.contentAt(50, 75, 200, 400)
	if x != 100 || y != 300 {
		t.Fatalf("fitted pointer = %.1f, %.1f", x, y)
	}
	if s.allowPageScroll() {
		t.Fatal("locked view allows scrolling")
	}
	s.pullScroll()
	if s.scrollY != 0 {
		t.Fatal("locked view retained its scroll offset")
	}
}

// TestViewSizeShrinksForDock checks the page viewport follows the dock.
func TestViewSizeShrinksForDock(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.screenW, s.screenH = 800, 600
	s.dev.on = false

	if w, h := s.viewSize(); w != 800 || h != 600 {
		t.Fatalf("closed dock view = %dx%d, want 800x600", w, h)
	}

	s.dev.on = true
	// An 800 px window gives the panel 224 px, not the full 340.
	if w, h := s.viewSize(); w != 576 || h != 600 {
		t.Fatalf("open dock view = %dx%d, want 576x600", w, h)
	}
}

// TestResizeShrinksPageForDock checks opening and closing the dock commits
// the smaller and then the full page size.
func TestResizeShrinksPageForDock(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	d.width, d.height = 800, 600
	s := newDevShell(d)
	s.pendingW, s.pendingH = 800, 600
	s.screenW, s.screenH = 800, 600
	s.dev.on = true

	if err := s.resize(); err != nil {
		t.Fatal(err)
	}

	if w, h := d.Size(); w != 576 || h != 600 {
		t.Fatalf("open dock page = %dx%d, want 576x600", w, h)
	}

	s.dev.on = false
	if err := s.resize(); err != nil {
		t.Fatal(err)
	}

	if w, h := d.Size(); w != 800 || h != 600 {
		t.Fatalf("closed dock page = %dx%d, want 800x600", w, h)
	}
}

// TestMinClampedPageScrolls checks a page that cannot shrink to the viewport
// is not scaled, so a pointer stays on its pixel.
func TestMinClampedPageScrolls(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	d.width, d.height = 1440, 960
	d.minW = 1120
	s := newDevShell(d)
	s.pendingW, s.pendingH = 1440, 960
	s.screenW, s.screenH = 1440, 960
	s.dev.on = true

	if err := s.resize(); err != nil {
		t.Fatal(err)
	}

	if s.stretched() {
		t.Fatal("a min-clamped page must scroll, not stretch")
	}

	frameW, frameH := s.frameSize()
	if x, _ := s.contentAt(100, 50, frameW, frameH); x != 100 {
		t.Fatalf("pointer x = %v, want 100", x)
	}
}

func TestResponsiveViewportBelowDesktopMinimum(t *testing.T) {
	ctx := context.Background()
	p, err := page.New(page.Config{
		HTML:  `<style>body{margin:0}header{display:flex;justify-content:flex-end;width:100%;height:40px}button{width:40px;height:40px}#compact{display:none}@media(max-width:700px){#compact{display:block}}</style><header><button id="close">X</button></header><div id="compact">Compact</div>`,
		Width: 1120, Height: 760, MinWidth: 900, MinHeight: 610,
	})
	if err != nil {
		t.Fatal(err)
	}
	s := newShell(ctx, p)
	defer s.cancel()
	for _, width := range []int{600, 1200, 600} {
		s.Layout(width, 700)
		s.moving = false
		if err := s.resize(); err != nil {
			t.Fatal(err)
		}
		if w, h := p.Size(); w != width || h != 700 {
			t.Fatalf("viewport=%dx%d, want %dx700", w, h, width)
		}
		compact := false
		closeFound := false
		for _, b := range p.Boxes() {
			if b.ID == "compact" && b.W > 0 && b.H > 0 {
				compact = true
			}
			if b.ID == "close" {
				closeFound = true
				if b.X+b.W > float64(width) {
					t.Fatalf("close extends past viewport: %+v", b)
				}
			}
		}
		if !closeFound {
			t.Fatal("missing close control")
		}
		if compact != (width <= 700) {
			t.Fatalf("width=%d compact media query active=%v", width, compact)
		}
	}
	if w, _ := p.MinSize(); w != 900 {
		t.Fatalf("native minimum changed to %d", w)
	}
}

// TestMain fails the binary when window tests leak goroutines (GEN-26).
func TestMain(m *testing.M) {
	clipboard.UseMemory(true)
	// Desktop tile tests use retained surfaces; GPU sessions can select clearing explicitly.
	ebiten.SetScreenClearedEveryFrame(false)
	// Unit tests must not claim or query the developer's real accessibility bus.
	_ = os.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	os.Exit(leakcheck.Verify(m))
}
