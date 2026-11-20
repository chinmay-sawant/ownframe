package window

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"os"
	"testing"
)

func TestScrollClampsAfterARedraw(t *testing.T) {
	t.Parallel()

	app := &scrollScreen{fakeScreen: fakeScreen{width: 600, height: 300, minW: 1, minH: 1}}
	app.boxes = []layout.Box{{ID: "tall", X: 0, Y: 0, W: 600, H: 1200}}

	shell := newShell(context.Background(), app)

	if err := shell.syncImage(); err != nil {
		t.Fatal(err)
	}

	shell.scrollY = 900

	// A Load, Back, or Forward redraws at the same size with new content.
	app.redraws++
	app.boxes = nil

	if err := shell.syncImage(); err != nil {
		t.Fatal(err)
	}

	if shell.scrollY != 0 {
		t.Fatalf("scrollY = %d after the redraw, want 0", shell.scrollY)
	}
}

// requestScreen is a fakeScreen with one pending scroll request.
type requestScreen struct {
	*fakeScreen
	req host.Scroll
	has bool
}

func (f *requestScreen) TakeScroll() (host.Scroll, bool) {
	if !f.has {
		return host.Scroll{}, false
	}

	f.has = false

	return f.req, true
}

func TestScrollRequestClamps(t *testing.T) {
	t.Parallel()

	app := &requestScreen{
		fakeScreen: &fakeScreen{
			width: 200, height: 200,
			boxes: []layout.Box{{X: 0, Y: 0, W: 400, H: 500}},
		},
		req: host.Scroll{Y: 1000, Absolute: true},
		has: true,
	}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}

	s.applyScrollRequest()

	if s.scrollX != 0 || s.scrollY != 300 {
		t.Fatalf("absolute = %d, %d", s.scrollX, s.scrollY)
	}

	app.req = host.Scroll{X: 10, Y: -40}
	app.has = true
	s.applyScrollRequest()

	if s.scrollX != 10 || s.scrollY != 260 {
		t.Fatalf("delta = %d, %d", s.scrollX, s.scrollY)
	}

	s.applyScrollRequest()

	if s.scrollY != 260 {
		t.Fatalf("cleared request moved to %d", s.scrollY)
	}
}

func TestScrollRequestPlainScreen(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{}, ctx: context.Background()}
	s.applyScrollRequest()
}

type stepScreen struct {
	*windowScreen
	changed bool
	calls   int
}

func (s *stepScreen) StepScrollWindow() bool { s.calls++; return s.changed }

func TestScrollReusesUnchangedWindow(t *testing.T) {
	app := &stepScreen{windowScreen: &windowScreen{fakeScreen: &fakeScreen{}, on: true}}
	s := &shell{app: app, ctx: context.Background(), scrollY: 100}
	if err := s.syncScrollWindow(); err != nil {
		t.Fatal(err)
	}
	if app.redraws != 0 || app.calls != 1 || app.oy != 100 {
		t.Fatal("unchanged window was relaid out")
	}
	app.changed = true
	s.scrollY = 200
	if err := s.syncScrollWindow(); err != nil {
		t.Fatal(err)
	}
	if app.redraws != 1 {
		t.Fatal("changed row window did not redraw")
	}
}

// windowScreen is a fakeScreen that observes scroll offsets.
type windowScreen struct {
	*fakeScreen
	ox, oy int
	on     bool
}

func (f *windowScreen) SetScrollOffset(x, y int) { f.ox, f.oy = x, y }
func (f *windowScreen) ScrollOffset() (int, int) { return f.ox, f.oy }
func (f *windowScreen) Windowing() bool          { return f.on }

func TestSyncScrollWindow(t *testing.T) {
	t.Parallel()

	app := &windowScreen{fakeScreen: &fakeScreen{}, on: true}
	s := &shell{app: app, ctx: context.Background(), scrollX: 10, scrollY: 200}

	if err := s.syncScrollWindow(); err != nil {
		t.Fatal(err)
	}
	if app.ox != 10 || app.oy != 200 {
		t.Fatalf("offset = %d, %d, want 10, 200", app.ox, app.oy)
	}
	if app.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", app.redraws)
	}

	app.redraws = 0
	if err := s.syncScrollWindow(); err != nil {
		t.Fatal(err)
	}
	if app.redraws != 0 {
		t.Fatalf("redraws = %d on unchanged offset, want 0", app.redraws)
	}

	app.on = false
	s.scrollY = 300
	if err := s.syncScrollWindow(); err != nil {
		t.Fatal(err)
	}
	if app.oy != 300 {
		t.Fatalf("offset y = %d, want 300", app.oy)
	}
	if app.redraws != 0 {
		t.Fatalf("redraws = %d when opted out, want 0", app.redraws)
	}
}

func TestOversizedWithoutADeviceLimit(t *testing.T) {
	t.Parallel()

	if oversized(480, 640) {
		t.Fatal("normal canvas oversized without a device limit")
	}
}

func TestScrollbarThumb(t *testing.T) {
	t.Parallel()

	pos, length := scrollbarThumb(100, 200, 100, 0)
	if pos != 0 || length != 50 {
		t.Fatalf("start pos=%v len=%v", pos, length)
	}

	pos, _ = scrollbarThumb(100, 200, 100, 100)
	if pos != 50 {
		t.Fatalf("end pos=%v", pos)
	}

	_, length = scrollbarThumb(100, 10000, 100, 0)
	if length != scrollbarMinThumb {
		t.Fatalf("min length=%v", length)
	}

	for _, off := range []int{0, 33, 100} {
		pos, _ := scrollbarThumb(100, 200, 100, off)
		if got := scrollbarOffset(float64(pos), 100, 200, 100); got != off {
			t.Fatalf("offset %d pos=%v round trip = %d", off, pos, got)
		}
	}
}

func TestScrollbarVisible(t *testing.T) {
	t.Parallel()

	if !scrollbarVisible(200, 100) || scrollbarVisible(100, 100) || scrollbarVisible(50, 0) {
		t.Fatal("visible check")
	}
}

func TestScrollbarOffsetClamps(t *testing.T) {
	t.Parallel()

	if got := scrollbarOffset(-50, 100, 200, 100); got != 0 {
		t.Fatalf("negative offset = %d", got)
	}

	if got := scrollbarOffset(1000, 100, 200, 100); got != 100 {
		t.Fatalf("past-end offset = %d", got)
	}
}

func TestScrollbarThumbNearlyFull(t *testing.T) {
	t.Parallel()

	// Content barely taller than the viewport: the thumb fills the track.
	pos, length := scrollbarThumb(100, 101, 100, 1)
	if length < 99 || length > 100 {
		t.Fatalf("length = %v", length)
	}

	if pos > 1.5 {
		t.Fatalf("pos = %v", pos)
	}
}

func TestContentSizeIncludesOverflow(t *testing.T) {
	t.Parallel()

	s := &shell{
		app: &fakeScreen{
			width: 320, height: 400,
			boxes: []layout.Box{{X: 0, Y: 0, W: 500, H: 700}},
		},
		screenW: 320, screenH: 400,
	}

	w, h := s.contentSize()
	if w != 500 || h != 700 {
		t.Fatalf("content = %dx%d", w, h)
	}
}

func TestThumbShowsResizeCursor(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 200, height: 200, minW: 1, minH: 1,
		boxes: []layout.Box{{W: 500, H: 500}},
	}
	s := newShell(context.Background(), app)

	var got []ebiten.CursorShapeType
	s.setCursor = func(shape ebiten.CursorShapeType) { got = append(got, shape) }

	s.pointerScrollbar(198, 10, false, false)
	s.pointerScrollbar(10, 198, false, false)
	s.pointerScrollbar(100, 100, false, false)

	want := []ebiten.CursorShapeType{ebiten.CursorShapeNSResize, ebiten.CursorShapeEWResize}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("cursor shapes = %v, want %v", got, want)
	}
}

func TestDrawPaintsVerticalThumb(t *testing.T) {
	if os.Getenv("OWNFRAME_SCROLLBAR_GPU_TEST") != "1" {
		t.Skip("requires a graphics session; set OWNFRAME_SCROLLBAR_GPU_TEST=1")
	}
	s := newShell(context.Background(), &fakeScreen{width: 100, height: 100})
	s.display = &layout.Display{Width: 100, Height: 200, PointsPerPixel: 1}
	ebiten.SetWindowSize(100, 100)
	g := &scrollbarDrawGame{shell: s}
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
}

type scrollbarDrawGame struct {
	shell *shell
	err   error
	drawn bool
}

func (g *scrollbarDrawGame) Layout(int, int) (int, int) { return 100, 100 }
func (g *scrollbarDrawGame) Update() error {
	if g.drawn {
		return ebiten.Termination
	}
	return nil
}
func (g *scrollbarDrawGame) Draw(screen *ebiten.Image) {
	if g.drawn {
		return
	}
	g.shell.Draw(screen)
	pixels := make([]byte, 100*100*4)
	screen.ReadPixels(pixels)
	at := (10*100 + 95) * 4
	want := []byte{scrollThumb.R, scrollThumb.G, scrollThumb.B, scrollThumb.A}
	for i := range want {
		if pixels[at+i] != want[i] { //nolint:gosec // reason: at+i bounded by the fixed 100x100x4 buffer.
			g.err = fmt.Errorf("thumb pixel = %v, want %v", pixels[at:at+4], want)
			break
		}
	}
	g.drawn = true
}

func TestThumbHoverKeepsResizeCursor(t *testing.T) {
	t.Parallel()

	app := &cursorScreen{fakeScreen: &fakeScreen{
		width: 200, height: 200, minW: 1, minH: 1,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := newShell(context.Background(), app)

	var got []ebiten.CursorShapeType
	s.setCursor = func(shape ebiten.CursorShapeType) { got = append(got, shape) }

	frameW, frameH := s.frameSize()
	thumbs := []struct {
		x, y int
		want ebiten.CursorShapeType
	}{
		{198, 10, ebiten.CursorShapeNSResize},
		{10, 198, ebiten.CursorShapeEWResize},
	}

	for _, tc := range thumbs {
		got = nil
		s.pointerScrollbar(tc.x, tc.y, false, false)
		px, py := s.contentAt(tc.x, tc.y, frameW, frameH)
		if err := s.hoverCursor(tc.x, tc.y, px, py); err != nil {
			t.Fatal(err)
		}
		if s.cursor != tc.want {
			t.Fatalf("hover at %d,%d cursor = %v, want %v", tc.x, tc.y, s.cursor, tc.want)
		}
		if len(got) != 1 || got[0] != tc.want {
			t.Fatalf("hover at %d,%d shapes = %v, want [%v]", tc.x, tc.y, got, tc.want)
		}
	}

	got = nil
	px, py := s.contentAt(100, 100, frameW, frameH)
	if err := s.hoverCursor(100, 100, px, py); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeDefault {
		t.Fatalf("off-thumb cursor = %v, want default", s.cursor)
	}
	if len(got) != 1 || got[0] != ebiten.CursorShapeDefault {
		t.Fatalf("off-thumb shapes = %v, want [default]", got)
	}

	s.dragAxis = dragHorizontal
	got = nil
	px, py = s.contentAt(198, 10, frameW, frameH)
	if err := s.hoverCursor(198, 10, px, py); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeEWResize {
		t.Fatalf("cursor during horizontal drag = %v, want %v", s.cursor, ebiten.CursorShapeEWResize)
	}
}

type scrollbarModeScreen struct {
	*fakeScreen
	mode host.ScrollbarMode
}

func (s *scrollbarModeScreen) ScrollbarMode() host.ScrollbarMode { return s.mode }

func TestScrollbarModePlatformDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode host.ScrollbarMode
		goos string
		want bool
	}{
		{name: "desktop default", mode: host.ScrollbarAuto, goos: "windows", want: true},
		{name: "web default", mode: host.ScrollbarAuto, goos: "js", want: true},
		{name: "android default", mode: host.ScrollbarAuto, goos: "android"},
		{name: "ios default", mode: host.ScrollbarAuto, goos: "ios"},
		{name: "show on android", mode: host.ScrollbarVisible, goos: "android", want: true},
		{name: "hide on desktop", mode: host.ScrollbarHidden, goos: "linux"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scrollbarModeEnabled(tc.mode, tc.goos); got != tc.want {
				t.Fatalf("scrollbarModeEnabled(%v, %q) = %t", tc.mode, tc.goos, got)
			}
		})
	}
}

func TestScrollbarGrabHonorsHiddenMode(t *testing.T) {
	t.Parallel()

	s := &shell{
		app: &scrollbarModeScreen{
			fakeScreen: &fakeScreen{
				width: 100, height: 100,
				boxes: []layout.Box{{X: 0, Y: 0, W: 100, H: 200}},
			},
			mode: host.ScrollbarHidden,
		},
		screenW: 100, screenH: 100,
	}

	if _, _, ok := s.scrollbarGrab(95, 10); ok {
		t.Fatal("hidden scrollbar was hit-testable")
	}
}
