package window

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"slices"
	"testing"
	"time"
)

// cursorScreen is a fakeScreen with a scripted hover shape.
type cursorScreen struct {
	*fakeScreen
	shape host.Shape
}

func (f *cursorScreen) CursorShape() host.Shape { return f.shape }

func TestCursorShapeSequence(t *testing.T) {
	t.Parallel()

	app := &cursorScreen{fakeScreen: &fakeScreen{}}
	var calls []ebiten.CursorShapeType

	s := &shell{app: app, ctx: context.Background()}
	s.setCursor = func(shape ebiten.CursorShapeType) {
		calls = append(calls, shape)
	}

	for _, shape := range []host.Shape{host.ShapeText, host.ShapePointer, host.ShapeDefault} {
		app.shape = shape
		s.applyCursor()
	}

	want := []ebiten.CursorShapeType{
		ebiten.CursorShapeText,
		ebiten.CursorShapePointer,
		ebiten.CursorShapeDefault,
	}
	if !slices.Equal(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}

	s.applyCursor()

	if len(calls) != 3 {
		t.Fatalf("repeat calls = %d", len(calls))
	}

	if s.cursor != ebiten.CursorShapeDefault {
		t.Fatalf("cursor = %v", s.cursor)
	}
}

func TestCursorMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   host.Shape
		want ebiten.CursorShapeType
	}{
		{host.ShapeDefault, ebiten.CursorShapeDefault},
		{host.ShapeText, ebiten.CursorShapeText},
		{host.ShapePointer, ebiten.CursorShapePointer},
		{host.ShapeResizeEW, ebiten.CursorShapeEWResize},
		{host.ShapeResizeNS, ebiten.CursorShapeNSResize},
	}

	for _, c := range cases {
		if got := cursorShape(c.in); got != c.want {
			t.Fatalf("cursorShape(%v) = %v", c.in, got)
		}
	}
}

func TestCursorPlainScreen(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{}, ctx: context.Background()}
	s.applyCursor()

	if s.cursor != ebiten.CursorShapeDefault {
		t.Fatalf("cursor = %v", s.cursor)
	}
}

// TestPressLeavesHandlerFocus pins the order: the caret lands before the
// click handler runs, so the handler's focus and selection survive.
func TestPressLeavesHandlerFocus(t *testing.T) {
	t.Parallel()

	app := &selectScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: app, ctx: context.Background()}

	if err := s.pressAt(10, 10); err != nil {
		t.Fatal(err)
	}

	if !app.focus {
		t.Fatal("SelectAt cleared the focus the click handler set")
	}
}

func TestClickWatchCounts(t *testing.T) {
	t.Parallel()

	var w clickWatch
	t0 := time.Unix(0, 0)

	if got := w.step(t0, 10, 10); got != 1 {
		t.Fatalf("first = %d", got)
	}

	if got := w.step(t0.Add(100*time.Millisecond), 11, 11); got != 2 {
		t.Fatalf("double = %d", got)
	}

	if got := w.step(t0.Add(200*time.Millisecond), 11, 12); got != 3 {
		t.Fatalf("triple = %d", got)
	}

	if got := w.step(t0.Add(300*time.Millisecond), 40, 40); got != 1 {
		t.Fatalf("far point = %d", got)
	}

	if got := w.step(t0.Add(900*time.Millisecond), 40, 40); got != 1 {
		t.Fatalf("slow click = %d", got)
	}
}

func TestEdgeScroll(t *testing.T) {
	t.Parallel()

	cases := []struct {
		y    int
		want int
	}{{-1, -48}, {0, 0}, {199, 0}, {200, 48}, {400, 48}}

	for _, c := range cases {
		if got := edgeScroll(c.y, 200, 48); got != c.want {
			t.Fatalf("edgeScroll(%d) = %d, want %d", c.y, got, c.want)
		}
	}
}

func TestDragScrollClamps(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{X: 0, Y: 0, W: 200, H: 500}},
	}
	s := &shell{app: app, screenW: 200, screenH: 200, dragActive: true}

	s.dragScroll(200)
	if s.scrollY != 48 {
		t.Fatalf("first step = %d", s.scrollY)
	}

	for i := 0; i < 20; i++ {
		s.dragScroll(200)
	}

	if s.scrollY != 300 {
		t.Fatalf("clamped = %d", s.scrollY)
	}

	s.dragScroll(-1)
	if s.scrollY != 252 {
		t.Fatalf("back up = %d", s.scrollY)
	}

	s.dragActive = false
	s.dragScroll(200)

	if s.scrollY != 252 {
		t.Fatalf("idle drag moved = %d", s.scrollY)
	}
}

func TestPressDragRelease(t *testing.T) {
	t.Parallel()

	app := &selectScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: app, ctx: context.Background()}

	if err := s.pressAt(10, 10); err != nil {
		t.Fatal(err)
	}

	if !s.dragActive {
		t.Fatal("drag not active after a press")
	}

	if err := s.dragAt(10, 10); err != nil {
		t.Fatal(err)
	}

	if err := s.dragAt(20, 30); err != nil {
		t.Fatal(err)
	}

	if err := s.releaseAt(); err != nil {
		t.Fatal(err)
	}

	if s.dragActive {
		t.Fatal("drag active after the release")
	}

	want := []string{"press", "at", "click", "drag", "release"}
	if !slices.Equal(app.calls, want) {
		t.Fatalf("calls = %v, want %v", app.calls, want)
	}
}

func TestDoubleAndTripleClick(t *testing.T) {
	t.Parallel()

	app := &selectScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: app, ctx: context.Background()}

	for i := 0; i < 3; i++ {
		if err := s.pressAt(10+float64(i), 10); err != nil {
			t.Fatal(err)
		}
	}

	want := []string{
		"press", "at", "click",
		"press", "at", "click", "word",
		"press", "at", "click", "line",
	}

	if !slices.Equal(app.calls, want) {
		t.Fatalf("calls = %v, want %v", app.calls, want)
	}
}

func TestPlainScreenKeepsClicks(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{}, ctx: context.Background()}

	if err := s.pressAt(1, 2); err != nil {
		t.Fatal(err)
	}

	if s.dragActive {
		t.Fatal("drag active on a plain screen")
	}

	if err := s.dragAt(3, 4); err != nil {
		t.Fatal(err)
	}

	if err := s.releaseAt(); err != nil {
		t.Fatal(err)
	}
}

func TestClaimedHoldDragsWithoutScrolling(t *testing.T) {
	t.Parallel()

	app := &holdScreen{
		selectScreen: &selectScreen{fakeScreen: &fakeScreen{width: 200, height: 200}},
		claim:        true,
	}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}
	t0 := time.Unix(0, 0)

	s.hold.arm(t0, 10, 10)
	if err := s.holdCheck(t0.Add(longPressDelay)); err != nil {
		t.Fatal(err)
	}

	if !s.hold.claimed || !s.dragActive {
		t.Fatalf("claimed = %v, dragActive = %v", s.hold.claimed, s.dragActive)
	}

	s.scrollY = 40
	now := []touchPos{{id: 1, x: 10, y: 60}}

	if err := s.touchMove(touchUpdate{dy: 50}, now, 200, 200); err != nil {
		t.Fatal(err)
	}

	if s.scrollY != 40 {
		t.Fatalf("scrollY = %d, want 40", s.scrollY)
	}

	if !slices.Contains(app.calls, "drag") {
		t.Fatalf("calls = %v, want a drag", app.calls)
	}
}

func TestUnclaimedTouchMoveScrolls(t *testing.T) {
	t.Parallel()

	app := &holdScreen{selectScreen: &selectScreen{fakeScreen: &fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{X: 0, Y: 0, W: 200, H: 500}},
	}}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}

	now := []touchPos{{id: 1, x: 10, y: 60}}

	if err := s.touchMove(touchUpdate{dy: -50}, now, 200, 200); err != nil {
		t.Fatal(err)
	}

	if s.scrollY != 50 {
		t.Fatalf("scrollY = %d, want 50", s.scrollY)
	}

	if slices.Contains(app.calls, "drag") {
		t.Fatalf("calls = %v, want no drag", app.calls)
	}
}

// hoverScreen counts hover passes and moves the generation like a redraw.
type hoverScreen struct {
	*fakeScreen
	display *layout.Display
	gen     uint64
	hovers  int
}

func (h *hoverScreen) Display() *layout.Display { return h.display }
func (h *hoverScreen) Generation() uint64       { return h.gen }
func (h *hoverScreen) Hover(context.Context, float64, float64) error {
	h.hovers++

	return nil
}

func newHoverShell(gen uint64) (*shell, *hoverScreen) {
	app := &hoverScreen{
		fakeScreen: &fakeScreen{width: 200, height: 200, minW: 1, minH: 1},
		display:    &layout.Display{Width: 200, Height: 200, PointsPerPixel: 1},
		gen:        gen,
	}
	s := &shell{
		app: app, ctx: context.Background(),
		display: app.display, seq: gen,
		screenW: 200, screenH: 200,
	}

	return s, app
}

func TestHoverStillPointerSkipsUntilRedraw(t *testing.T) {
	t.Parallel()

	s, app := newHoverShell(1)
	s.frameCommit()

	for i := 0; i < 2; i++ {
		if err := s.hoverCursor(50, 50, 50, 50); err != nil {
			t.Fatal(err)
		}
	}
	if app.hovers != 1 {
		t.Fatalf("still pointer calls = %d, want 1", app.hovers)
	}

	// A redraw moves the generation and the shell rebuilds, so the next
	// frame resolves hover again.
	app.gen++
	app.display = &layout.Display{Width: 200, Height: 200, PointsPerPixel: 1}
	if err := s.syncImage(); err != nil {
		t.Fatal(err)
	}
	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 2 {
		t.Fatalf("post-redraw calls = %d, want 2", app.hovers)
	}
}

func TestHoverMovedPointResolves(t *testing.T) {
	t.Parallel()

	s, app := newHoverShell(1)
	s.frameCommit()

	points := [][2]float64{{50, 50}, {60, 60}, {60, 108}}
	for _, p := range points {
		if err := s.hoverCursor(int(p[0]), int(p[1]), p[0], p[1]); err != nil {
			t.Fatal(err)
		}
	}
	if app.hovers != 3 {
		t.Fatalf("moving point calls = %d, want 3", app.hovers)
	}
}

// TestHoverDeferredUntilFirstFrame keeps the first update from pushing the
// cursor hover to the page, so startup paints without a hover redraw. The
// skip leaves the hover due, so the update after the first committed frame
// still resolves it.
func TestHoverDeferredUntilFirstFrame(t *testing.T) {
	t.Parallel()

	s, app := newHoverShell(1)

	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 0 {
		t.Fatalf("pre-frame calls = %d, want 0", app.hovers)
	}

	s.frameCommit()
	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 1 {
		t.Fatalf("post-frame calls = %d, want 1", app.hovers)
	}
}

// pageHoverScreen reports the element it paints as hovered, so the shell can
// tell its own hover from one the page set.
type pageHoverScreen struct {
	*fakeScreen
	display *layout.Display
	gen     uint64
	id      string
	next    string
	hovers  int
}

func (h *pageHoverScreen) Display() *layout.Display { return h.display }
func (h *pageHoverScreen) Generation() uint64       { return h.gen }
func (h *pageHoverScreen) HoverID() string          { return h.id }

// Hover repaints the element at the point and moves the generation, the way
// a real page redraw does.
func (h *pageHoverScreen) Hover(context.Context, float64, float64) error {
	h.hovers++
	if h.next != "" && h.next != h.id {
		h.id = h.next
		h.moveGeneration()
	}

	return nil
}

func (h *pageHoverScreen) moveGeneration() {
	h.gen++
	h.display = &layout.Display{Width: 200, Height: 200, PointsPerPixel: 1}
}

func newPageHoverShell(gen uint64) (*shell, *pageHoverScreen) {
	app := &pageHoverScreen{
		fakeScreen: &fakeScreen{width: 200, height: 200, minW: 1, minH: 1},
		display:    &layout.Display{Width: 200, Height: 200, PointsPerPixel: 1},
		gen:        gen,
	}
	s := &shell{
		app: app, ctx: context.Background(),
		display: app.display, seq: gen,
		screenW: 200, screenH: 200,
	}

	return s, app
}

// TestHoverOwnRedrawSkipsRepeat checks that a hover push which redrew the
// page does not make the same point resolve again on the next frame. The
// generation moved because of the push itself, not because a box under the
// cursor moved.
func TestHoverOwnRedrawSkipsRepeat(t *testing.T) {
	t.Parallel()

	s, app := newPageHoverShell(1)
	s.frameCommit()
	app.next = "row-a"

	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 1 {
		t.Fatalf("first push calls = %d, want 1", app.hovers)
	}

	if err := s.syncImage(); err != nil {
		t.Fatal(err)
	}
	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 1 {
		t.Fatalf("repeat after own redraw = %d, want 1", app.hovers)
	}
}

// TestHoverPageHoverSurvives checks that a hover the page set itself stays
// until the pointer moves. The shell's stale point must not be pushed back
// over it, because that costs the page a second layout for one frame.
func TestHoverPageHoverSurvives(t *testing.T) {
	t.Parallel()

	s, app := newPageHoverShell(1)
	s.frameCommit()
	app.next = "row-a"

	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 1 {
		t.Fatalf("first push calls = %d, want 1", app.hovers)
	}
	if err := s.syncImage(); err != nil {
		t.Fatal(err)
	}

	// A programmatic Hover moved the page's hover and redrew it.
	app.id = "row-b"
	app.moveGeneration()
	if err := s.syncImage(); err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if err := s.hoverCursor(50, 50, 50, 50); err != nil {
			t.Fatal(err)
		}
	}
	if app.hovers != 1 {
		t.Fatalf("still pointer overridden the page hover: calls = %d, want 1", app.hovers)
	}

	// A moved pointer takes hover back.
	if err := s.hoverCursor(60, 60, 60, 60); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 2 {
		t.Fatalf("moved pointer calls = %d, want 2", app.hovers)
	}
}

// TestHoverDataRedrawResolves checks that a generation change with the
// window's own hover still painted resolves the point again, so a box that
// moved under a still pointer is followed.
func TestHoverDataRedrawResolves(t *testing.T) {
	t.Parallel()

	s, app := newPageHoverShell(1)
	s.frameCommit()
	app.next = "row-a"

	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if err := s.syncImage(); err != nil {
		t.Fatal(err)
	}

	app.moveGeneration()
	if err := s.syncImage(); err != nil {
		t.Fatal(err)
	}
	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 2 {
		t.Fatalf("post-data-redraw calls = %d, want 2", app.hovers)
	}
}

// holdScreen records the long-press calls the window makes. claim is the
// answer its LongPress returns.
type holdScreen struct {
	*selectScreen
	claim bool
	holds int
	x, y  float64
}

func (f *holdScreen) LongPress(_ context.Context, x, y float64) (bool, error) {
	f.holds++
	f.x, f.y = x, y

	return f.claim, nil
}

func TestLongPressFiresOnce(t *testing.T) {
	t.Parallel()

	app := &holdScreen{selectScreen: &selectScreen{fakeScreen: &fakeScreen{}}}
	s := &shell{app: app, ctx: context.Background()}
	t0 := time.Unix(0, 0)

	s.hold.arm(t0, 10, 20)

	if err := s.holdCheck(t0.Add(longPressDelay - time.Millisecond)); err != nil {
		t.Fatal(err)
	}

	if app.holds != 0 {
		t.Fatalf("holds = %d before the delay", app.holds)
	}

	if err := s.holdCheck(t0.Add(longPressDelay)); err != nil {
		t.Fatal(err)
	}

	if app.holds != 1 || app.x != 10 || app.y != 20 {
		t.Fatalf("holds = %d at %v, %v", app.holds, app.x, app.y)
	}

	if err := s.holdCheck(t0.Add(2 * longPressDelay)); err != nil {
		t.Fatal(err)
	}

	if app.holds != 1 {
		t.Fatalf("holds = %d after a repeat", app.holds)
	}

	s.hold.arm(t0, 10, 20)
	s.hold.release()

	if err := s.holdCheck(t0.Add(longPressDelay)); err != nil {
		t.Fatal(err)
	}

	if app.holds != 1 {
		t.Fatalf("holds = %d after a release", app.holds)
	}
}

func TestLongPressSlopCancels(t *testing.T) {
	t.Parallel()

	app := &holdScreen{selectScreen: &selectScreen{fakeScreen: &fakeScreen{}}}
	s := &shell{app: app, ctx: context.Background()}
	t0 := time.Unix(0, 0)

	s.hold.arm(t0, 10, 20)
	s.hold.move(14, 22)

	if err := s.holdCheck(t0.Add(longPressDelay)); err != nil {
		t.Fatal(err)
	}

	if app.holds != 1 {
		t.Fatalf("a small move canceled the hold, holds = %d", app.holds)
	}

	s.hold.arm(t0, 10, 20)
	s.hold.move(19, 20)

	if err := s.holdCheck(t0.Add(longPressDelay)); err != nil {
		t.Fatal(err)
	}

	if app.holds != 1 {
		t.Fatalf("a moved press fired, holds = %d", app.holds)
	}
}

func TestPressedNow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		down bool
		was  bool
		want bool
	}{
		{false, false, false},
		{true, false, true},
		{true, true, false},
		{false, true, false},
	}

	for _, c := range cases {
		if got := pressedNow(c.down, c.was); got != c.want {
			t.Fatalf("pressedNow(%v, %v) = %v", c.down, c.was, got)
		}
	}
}

func TestFreshTouches(t *testing.T) {
	t.Parallel()

	got := freshTouches(
		[]ebiten.TouchID{1, 2, 5},
		[]ebiten.TouchID{2, 3},
	)
	if len(got) != 2 || got[0] != 1 || got[1] != 5 {
		t.Fatalf("fresh = %v", got)
	}

	if freshTouches(nil, []ebiten.TouchID{1}) != nil {
		t.Fatal("fresh from nil is not nil")
	}
}

func TestHeldPressDoesNotStartThumbDrag(t *testing.T) {
	app := &pointerRouteScreen{fakeScreen: fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200, mouseDown: true}
	if _, err := s.pointerMouse(198, 10, 198, 10, true, false, false); err != nil {
		t.Fatal(err)
	}
	if s.dragAxis != dragNone || app.clicks != 0 {
		t.Fatalf("held press grabbed thumb or clicked page: axis %d, clicks %d", s.dragAxis, app.clicks)
	}
}

func TestPressEdge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		down, released, was bool
		want                bool
	}{
		{"fresh press", true, false, false, true},
		{"held press", true, false, true, false},
		{"fast pair between frames", false, true, false, true},
		{"release after seen press", false, true, true, false},
		{"idle frame", false, false, false, false},
	}

	for _, c := range cases {
		if got := pressEdge(c.down, c.released, c.was); got != c.want {
			t.Fatalf("%s: pressEdge(%v, %v, %v) = %v", c.name,
				c.down, c.released, c.was, got)
		}
	}
}

func TestFastClickBetweenFramesStillClicks(t *testing.T) {
	app := &pointerRouteScreen{fakeScreen: fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}

	clicked, err := s.pointerMouse(100, 100, 100, 100, false, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !clicked || app.presses != 1 || app.clicks != 1 || app.releases != 1 {
		t.Fatalf("events = %v/%d/%d/%d, want one press, click, release", clicked,
			app.presses, app.clicks, app.releases)
	}
}

func TestReleaseAfterSeenPressDoesNotClickAgain(t *testing.T) {
	app := &pointerRouteScreen{fakeScreen: fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}

	if _, err := s.pointerMouse(100, 100, 100, 100, true, false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.pointerMouse(100, 100, 100, 100, false, true, false); err != nil {
		t.Fatal(err)
	}
	if app.clicks != 1 {
		t.Fatalf("clicks = %d, want 1", app.clicks)
	}
}

func TestContentClickAndScrollbarDragKeepTheirEvents(t *testing.T) {
	app := &pointerRouteScreen{fakeScreen: fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}
	move := func(x, y int, down, released bool) {
		t.Helper()
		if _, err := s.pointerMouse(x, y, float64(x), float64(y), down, released, false); err != nil {
			t.Fatal(err)
		}
	}

	move(100, 100, true, false)
	move(100, 100, false, true)
	if app.presses != 1 || app.clicks != 1 || app.releases != 1 {
		t.Fatalf("content click events = %d/%d/%d, want 1/1/1", app.presses, app.clicks, app.releases)
	}

	move(198, 10, true, false)
	move(198, 100, true, false)
	if s.scrollY == 0 || app.presses != 1 || app.clicks != 1 || app.drags != 0 {
		t.Fatalf("thumb drag offset/events = %d/%d/%d/%d", s.scrollY, app.presses, app.clicks, app.drags)
	}
	move(198, 100, false, true)
	if app.clicks != 1 || app.releases != 2 {
		t.Fatalf("thumb release events = clicks %d, releases %d", app.clicks, app.releases)
	}
	move(10, 198, true, false)
	move(100, 198, true, false)
	if s.scrollX == 0 || app.presses != 1 || app.clicks != 1 || app.drags != 0 {
		t.Fatalf("horizontal thumb offset/events = %d/%d/%d/%d", s.scrollX, app.presses, app.clicks, app.drags)
	}
	move(100, 198, false, true)

	_, wheelY := panScroll(0, s.scrollY, 0, -1, 500, 500, 200, 200)
	if wheelY <= s.scrollY {
		t.Fatalf("page wheel offset = %d after thumb offset %d", wheelY, s.scrollY)
	}
}

// TestTapAtPlacesCaret checks a tap lands the caret before the click
// handler runs, matching pressAt.
func TestTapAtPlacesCaret(t *testing.T) {
	t.Parallel()

	app := &selectScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: app, ctx: context.Background()}

	if err := s.tapAt(10, 10); err != nil {
		t.Fatal(err)
	}

	want := []string{"press", "at", "click", "release"}
	if !slices.Equal(app.calls, want) {
		t.Fatalf("calls = %v, want %v", app.calls, want)
	}
}

// TestTapAtPlainScreen checks a page without the selector keeps the old
// press, click, release tap.
func TestTapAtPlainScreen(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{}, ctx: context.Background()}

	if err := s.tapAt(1, 2); err != nil {
		t.Fatal(err)
	}
}

// TestDevPickAltPlacesCaret checks Alt+click lands the caret before the
// click handler runs.
func TestDevPickAltPlacesCaret(t *testing.T) {
	t.Parallel()

	app := &selectScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: app, ctx: context.Background()}

	if err := s.devPick(10, 10, true); err != nil {
		t.Fatal(err)
	}

	want := []string{"press", "at", "click"}
	if !slices.Equal(app.calls, want) {
		t.Fatalf("calls = %v, want %v", app.calls, want)
	}

	if !s.dev.forward {
		t.Fatal("release not forwarded after an alt pick")
	}
}

// TestDevPickAltPlainScreen checks Alt+click still forwards on a page
// without the selector.
func TestDevPickAltPlainScreen(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{}, ctx: context.Background()}

	if err := s.devPick(1, 2, true); err != nil {
		t.Fatal(err)
	}

	if !s.dev.forward {
		t.Fatal("release not forwarded after an alt pick")
	}
}
