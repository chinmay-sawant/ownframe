package window

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
	"testing"
	"time"
)

type cancelScreen struct {
	touchPressScreen
	cancel bool
}

func (f *cancelScreen) TakeTouchCancel() bool { want := f.cancel; f.cancel = false; return want }

func TestCancelledTouchCannotTapOrRestartWhileNativeIDsRemain(t *testing.T) {
	app := &cancelScreen{touchPressScreen: touchPressScreen{fakeScreen: fakeScreen{width: 200, height: 200}}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}
	if err := s.applyTouches([]touchPos{{id: 1, x: 20, y: 20}}, false, 200, 200); err != nil {
		t.Fatal(err)
	}
	app.cancel = true
	for _, ids := range [][]ebiten.TouchID{{1}, {1}, nil} {
		handled, err := s.consumeTouchCancel(ids)
		if err != nil || !handled {
			t.Fatalf("cancel %v %v", handled, err)
		}
	}
	if handled, _ := s.consumeTouchCancel(nil); handled {
		t.Fatal("cancellation never cleared")
	}
	if app.clicks != 0 || app.releases != 1 || s.hold.armed || len(s.fingers.fingers) != 0 {
		t.Fatal("cancel delivered tap or retained press")
	}
}

type captureScreen struct {
	fakeScreen
	claim       bool
	moves, ends int
}

func (f *captureScreen) BeginDrag(context.Context, float64, float64) (bool, error) {
	return f.claim, nil
}
func (f *captureScreen) MoveDrag(context.Context, float64, float64) error { f.moves++; return nil }
func (f *captureScreen) EndDrag(context.Context) error                    { f.ends++; return nil }

func TestTouchCaptureMovesBeforeLiftAndCancelsPinch(t *testing.T) {
	app := &captureScreen{fakeScreen: fakeScreen{width: 200, height: 200}, claim: true}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200, pageZoom: 1}
	finger := touchPos{id: 1, x: 40, y: 40}
	frames := [][]touchPos{{finger}, {{id: 1, x: 60, y: 45}}, {{id: 1, x: 60, y: 45}, {id: 2, x: 100, y: 100}}, nil}
	for i, now := range frames {
		u := s.fingers.frame(now, false)
		if _, err := s.routeTouchDrag(u, now, 200, 200); err != nil {
			t.Fatal(err)
		}
		if i == 1 && app.moves != 1 {
			t.Fatal("movement waits for lift")
		}
	}
	if app.ends != 1 || s.gesture.active || s.scrollY != 0 {
		t.Fatalf("capture ends=%d active=%v scroll=%d", app.ends, s.gesture.active, s.scrollY)
	}
}

func TestUnclaimedTouchRemainsScrollable(t *testing.T) {
	app := &captureScreen{fakeScreen: fakeScreen{width: 200, height: 200}, claim: false}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200, pageZoom: 1}
	p := touchPos{id: 1, x: 40, y: 40}
	u := s.fingers.frame([]touchPos{p}, false)
	handled, err := s.routeTouchDrag(u, []touchPos{p}, 200, 200)
	if err != nil || handled || s.gesture.active {
		t.Fatalf("unclaimed %v %v", handled, err)
	}
}

func TestDevToolsForwardsDragUntilRelease(t *testing.T) {
	app := &captureScreen{fakeScreen: fakeScreen{width: 200, height: 200}, claim: true}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}
	if err := s.devPick(40, 40, false); err != nil {
		t.Fatal(err)
	}
	if !s.dev.forward || !s.gesture.active {
		t.Fatal("content press did not start the captured drag")
	}
	if err := s.devForwardDrag(60, 50, true, false); err != nil {
		t.Fatal(err)
	}
	if err := s.devForwardDrag(60, 50, false, true); err != nil {
		t.Fatal(err)
	}
	if app.moves != 1 || app.ends != 1 || s.dev.forward || s.gesture.active {
		t.Fatalf("moves=%d ends=%d forward=%v active=%v", app.moves, app.ends, s.dev.forward, s.gesture.active)
	}
}

type touchReleaseScreen struct {
	fakeScreen
	releases int
}

func (f *touchReleaseScreen) Release(context.Context) error { f.releases++; return nil }

type noTouchZoomScreen struct{ fakeScreen }

func (noTouchZoomScreen) TouchZoomAllowed() bool { return false }

func TestTouchZoomPolicyPreservesOtherZoom(t *testing.T) {
	f := noTouchZoomScreen{fakeScreen: fakeScreen{width: 200, height: 200}}
	s := &shell{app: &f, ctx: context.Background(), screenW: 200, screenH: 200, pageZoom: 1.5}
	first := []touchPos{{id: 1, x: 20, y: 20}, {id: 2, x: 80, y: 20}}
	second := []touchPos{{id: 1, x: 10, y: 20}, {id: 2, x: 110, y: 20}}
	if err := s.applyTouches(first, false, 200, 200); err != nil {
		t.Fatal(err)
	}
	if err := s.applyTouches(second, false, 200, 200); err != nil {
		t.Fatal(err)
	}
	if s.fingers.zoomOr1() != 1 || s.zoom() != 1.5 {
		t.Fatalf("touch zoom %v, page zoom %v", s.fingers.zoomOr1(), s.zoom())
	}
}

func TestIdleTouchLeavesMousePressAlone(t *testing.T) {
	f := &touchReleaseScreen{fakeScreen: fakeScreen{width: 200, height: 200}}
	s := &shell{app: f, ctx: context.Background(), screenW: 200, screenH: 200, pageZoom: 1, mouseDown: true}
	if err := s.applyTouches(nil, true, 200, 200); err != nil {
		t.Fatal(err)
	}
	if f.releases != 0 {
		t.Fatal("idle touch released the mouse press")
	}
}

func TestMovingTouchReleasesOnLift(t *testing.T) {
	f := &touchReleaseScreen{fakeScreen: fakeScreen{width: 200, height: 200,
		boxes: []layout.Box{{W: 200, H: 400}}}}
	s := &shell{app: f, ctx: context.Background(), screenW: 200, screenH: 200, pageZoom: 1}
	for _, now := range [][]touchPos{{{id: 1, x: 40, y: 40}}, {{id: 1, x: 40, y: 20}}, nil} {
		if err := s.applyTouches(now, false, 200, 200); err != nil {
			t.Fatal(err)
		}
	}
	if f.releases != 2 || s.scrollY != 20 {
		t.Fatalf("releases %d scroll %d", f.releases, s.scrollY)
	}
}

func TestTouchDragScrolls(t *testing.T) {
	t.Parallel()

	var g touchGesture
	if u := g.frame([]touchPos{{id: 1, x: 10, y: 10}}, false); u.tap != nil {
		t.Fatal("tap on a fresh finger")
	}

	u := g.frame([]touchPos{{id: 1, x: 10, y: 60}}, false)
	if u.dx != 0 || u.dy != 50 || u.tap != nil {
		t.Fatalf("drag = %d, %d tap = %v", u.dx, u.dy, u.tap)
	}

	u = g.frame(nil, false)
	if u.tap != nil {
		t.Fatal("a moved finger tapped")
	}
}

func TestTouchTapLifts(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 10, y: 10}}, false)

	u := g.frame(nil, false)
	if u.tap == nil || u.tap.x != 10 || u.tap.y != 10 {
		t.Fatalf("tap = %v", u.tap)
	}
}

func TestTouchTapSuppressedByMouse(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 10, y: 10}}, true)

	u := g.frame(nil, false)
	if u.tap != nil {
		t.Fatal("a suppressed finger tapped")
	}
}

func TestTouchDragClampsAtBottom(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 0, y: 200}}, false)

	u := g.frame([]touchPos{{id: 1, x: 0, y: -200}}, false)

	sx, sy := clampScroll(0-u.dx, 0-u.dy, 200, 500, 200, 200)
	if sx != 0 || sy != 300 {
		t.Fatalf("offset = %d, %d", sx, sy)
	}
}

func TestTouchSlopKeepsATap(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 10, y: 10}}, false)

	u := g.frame([]touchPos{{id: 1, x: 14, y: 12}}, false)
	if u.dx != 0 || u.dy != 0 {
		t.Fatalf("slop moved = %d, %d", u.dx, u.dy)
	}

	u = g.frame(nil, false)
	if u.tap == nil {
		t.Fatal("a small move lost the tap")
	}
}

func TestTouchPinchSetsZoom(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 0, y: 0}, {id: 2, x: 100, y: 0}}, false)
	g.frame([]touchPos{{id: 1, x: 0, y: 0}, {id: 2, x: 150, y: 0}}, false)

	if got := g.zoomOr1(); got != 1.5 {
		t.Fatalf("zoom = %v", got)
	}

	u := g.frame(nil, false)
	if u.tap != nil {
		t.Fatal("pinch fingers tapped")
	}

	if got := g.zoomOr1(); got != 1.5 {
		t.Fatalf("zoom after release = %v", got)
	}
}

func TestPinchScale(t *testing.T) {
	t.Parallel()

	cases := []struct {
		start, now, base float64
		want             float64
	}{
		{100, 150, 1, 1.5},
		{100, 10, 1, minZoom},
		{100, 1000, 1, maxZoom},
		{0, 150, 1, 1},
		{100, 150, 2, 3},
	}

	for _, c := range cases {
		if got := pinchScale(c.start, c.now, c.base); got != c.want {
			t.Fatalf("pinchScale(%v, %v, %v) = %v", c.start, c.now, c.base, got)
		}
	}
}

func TestContentPointZoom(t *testing.T) {
	t.Parallel()

	// A box at content (50, 60) draws at screen (75, 90) under a 1.5 zoom.
	// A click there maps back to the box.
	x, y := contentPointZoom(75, 90, 0, 0, 1.5, false, 200, 200, 200, 200)
	if x != 50 || y != 60 {
		t.Fatalf("point = %v, %v", x, y)
	}

	s := &shell{
		app:     &fakeScreen{width: 200, height: 200},
		screenW: 200, screenH: 200, scrollX: 10, scrollY: 20,
	}
	s.fingers.zoom = 2

	x, y = s.contentAt(30, 40, 200, 200)
	if x != 20 || y != 30 {
		t.Fatalf("scrolled point = %v, %v", x, y)
	}
}

type touchPressScreen struct {
	fakeScreen
	presses, releases, clicks int
}

func (f *touchPressScreen) Press(context.Context, float64, float64) error { f.presses++; return nil }
func (f *touchPressScreen) Release(context.Context) error                 { f.releases++; return nil }
func (f *touchPressScreen) Click(context.Context, float64, float64) error { f.clicks++; return nil }

func TestTouchTapHasOnePressAndClick(t *testing.T) {
	f := &touchPressScreen{fakeScreen: fakeScreen{width: 200, height: 200}}
	s := &shell{app: f, ctx: context.Background(), screenW: 200, screenH: 200}
	for _, touches := range [][]touchPos{{{id: 1, x: 40, y: 40}}, nil} {
		if err := s.applyTouches(touches, false, 200, 200); err != nil {
			t.Fatal(err)
		}
	}
	if f.presses != 1 || f.clicks != 1 || f.releases != 1 {
		t.Fatalf("events: %+v", f)
	}
}

func TestPinchCancelsActivePressImmediately(t *testing.T) {
	f := &touchPressScreen{fakeScreen: fakeScreen{width: 200, height: 200}}
	s := &shell{app: f, ctx: context.Background(), screenW: 200, screenH: 200}
	for _, touches := range [][]touchPos{{{id: 1, x: 40, y: 40}}, {{id: 1, x: 40, y: 40}, {id: 2, x: 80, y: 40}}, nil} {
		if err := s.applyTouches(touches, false, 200, 200); err != nil {
			t.Fatal(err)
		}
	}
	if f.presses != 1 || f.clicks != 0 || s.hold.due(time.Now().Add(time.Second)) {
		t.Fatalf("events: %+v", f)
	}
}

func TestTouchScrollFlickStartsAtReducedVelocity(t *testing.T) {
	start := time.Unix(0, 0)
	m := newTouchScrollMotion(1, 5)
	m.begin(start)
	m.move(0, -200, start.Add(50*time.Millisecond))
	m.release()
	var distance int
	for i := 1; i <= 30; i++ {
		_, dy := m.step(start.Add(50*time.Millisecond + time.Duration(i)*time.Second/60))
		distance += dy
	}
	if distance < 250 || distance > 500 {
		t.Fatalf("quick flick distance = %d, want 250..500px", distance)
	}
}

func TestTouchScrollPauseBeforeReleaseCancelsFling(t *testing.T) {
	start := time.Unix(0, 0)
	m := newTouchScrollMotion(1, 5)
	m.begin(start)
	m.move(0, -40, start.Add(20*time.Millisecond))
	m.idle(start.Add(70 * time.Millisecond))
	m.release()
	_, dy := m.step(start.Add(90 * time.Millisecond))
	if dy != 0 {
		t.Fatalf("fling after stationary pause = %d, want 0", dy)
	}
}

func TestTouchScrollMotionFollowsFingerAndFlingDecays(t *testing.T) {
	start := time.Unix(0, 0)
	m := newTouchScrollMotion(1, 5)
	m.begin(start)
	dx, dy := m.move(0, -20, start.Add(20*time.Millisecond))
	if dx != 0 || dy != 20 {
		t.Fatalf("drag delta = %d, %d, want 0, 20", dx, dy)
	}
	m.release()
	_, first := m.step(start.Add(70 * time.Millisecond))
	_, second := m.step(start.Add(120 * time.Millisecond))
	if first <= 0 || second <= 0 || second >= first {
		t.Fatalf("fling deltas = %d, %d, want positive deceleration", first, second)
	}
}

func TestTouchScrollFlingIsUpdateRateIndependent(t *testing.T) {
	if a, b := flingDistance(60), flingDistance(90); math.Abs(a-b) > 2 {
		t.Fatalf("60 Hz distance %.2f, 90 Hz distance %.2f", a, b)
	}
}

func TestTouchScrollSensitivityScalesFingerMovement(t *testing.T) {
	start := time.Unix(0, 0)
	m := newTouchScrollMotion(.5, 5)
	m.begin(start)
	_, dy := m.move(0, -10, start.Add(20*time.Millisecond))
	if dy != 5 {
		t.Fatalf("scroll delta = %d, want 5", dy)
	}
}

func flingDistance(rate int) float64 {
	start := time.Unix(0, 0)
	m := newTouchScrollMotion(1, 5)
	m.begin(start)
	m.move(0, -200, start.Add(50*time.Millisecond))
	m.release()
	var distance float64
	for i := 1; i <= rate/2; i++ {
		_, dy := m.step(start.Add(50*time.Millisecond + time.Duration(i)*time.Second/time.Duration(rate)))
		distance += float64(dy)
	}
	return distance
}

func TestTouchUpdateReportsAFreshFinger(t *testing.T) {
	t.Parallel()

	var g touchGesture

	u := g.frame([]touchPos{{id: 7, x: 10, y: 20}}, false)
	if u.start == nil || u.start.id != 7 || u.start.x != 10 || u.start.y != 20 {
		t.Fatalf("start = %v", u.start)
	}

	u = g.frame([]touchPos{{id: 7, x: 30, y: 20}}, false)
	if u.start != nil {
		t.Fatalf("a moving finger restarted the hold: %v", u.start)
	}
}

func TestTouchSwipeUpLifts(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 10, y: 100}}, false)
	g.frame([]touchPos{{id: 1, x: 10, y: 40}}, false)

	u := g.frame(nil, false)
	if u.swipe == nil || u.swipe.dx != 0 || u.swipe.dy != -60 {
		t.Fatalf("swipe = %+v", u.swipe)
	}

	if u.tap != nil {
		t.Fatalf("tap = %+v", u.tap)
	}
}

func TestTouchSwipeDownLifts(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 10, y: 40}}, false)
	g.frame([]touchPos{{id: 1, x: 10, y: 100}}, false)

	u := g.frame(nil, false)
	if u.swipe == nil || u.swipe.dx != 0 || u.swipe.dy != 60 {
		t.Fatalf("swipe = %+v", u.swipe)
	}
}

func TestTouchSwipeNeedsDistance(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 10, y: 100}}, false)
	g.frame([]touchPos{{id: 1, x: 10, y: 88}}, false)

	u := g.frame(nil, false)
	if u.swipe != nil || u.tap != nil {
		t.Fatalf("swipe = %+v tap = %+v", u.swipe, u.tap)
	}
}

func TestTouchSwipeSuppressedByMulti(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 10, y: 100}, {id: 2, x: 100, y: 100}}, false)
	g.frame([]touchPos{{id: 1, x: 10, y: 40}, {id: 2, x: 100, y: 40}}, false)

	u := g.frame(nil, false)
	if u.swipe != nil || u.tap != nil {
		t.Fatalf("swipe = %+v tap = %+v", u.swipe, u.tap)
	}
}

func TestTouchSwipeSuppressedByEaten(t *testing.T) {
	t.Parallel()

	var g touchGesture
	g.frame([]touchPos{{id: 1, x: 10, y: 100}}, true)
	g.frame([]touchPos{{id: 1, x: 10, y: 40}}, false)

	u := g.frame(nil, false)
	if u.swipe != nil || u.tap != nil {
		t.Fatalf("swipe = %+v tap = %+v", u.swipe, u.tap)
	}
}

// TestTouchTapDisarmsHold checks a touch tap ends the held-press watch. A tap
// that left the watch armed fired the long press 450 ms later, after the
// finger was already up.
func TestTouchTapDisarmsHold(t *testing.T) {
	t.Parallel()

	app := &holdScreen{selectScreen: &selectScreen{fakeScreen: &fakeScreen{}}}
	s := &shell{app: app, ctx: context.Background()}
	t0 := time.Unix(0, 0)

	s.hold.arm(t0, 10, 20)

	if err := s.tapAt(10, 20); err != nil {
		t.Fatal(err)
	}

	if err := s.holdCheck(t0.Add(longPressDelay)); err != nil {
		t.Fatal(err)
	}

	if app.holds != 0 {
		t.Fatalf("holds = %d after a tap", app.holds)
	}
}
