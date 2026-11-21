package window

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"math"
	"runtime"
	"time"
)

// pointer sends hover and clicks. A click needs an up-to-down edge the shell
// saw itself. IsMouseButtonJustPressed can repeat for one press when a
// blocking call, such as the file dialog, stalls the event queue.
func (s *shell) pointer() error {
	if handled, err := s.consumeTouchCancel(ebiten.AppendTouchIDs(nil)); handled || err != nil {
		return err
	}
	x, y := s.cursorPosition()
	s.cursorX, s.cursorY = x, y
	if s.titlebarInset() > 0 && y < 0 && s.mouseDown {
		_, err := s.pointerMouse(x, y, -1, -1, ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft), inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft), true)
		return err
	}
	if s.titlebarInset() > 0 && y < 0 {
		if err := s.hoverCursor(x, y, -1, -1); err != nil {
			return err
		}
	}
	if s.titlebarPointer(x, y, ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)) {
		return nil
	}
	if s.titlebarInset() > 0 && y < 0 {
		_, err := s.moveWindow(x, y)
		return err
	}
	claimed, err := s.menuClaim(x, y,
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft))
	if err != nil {
		return err
	}
	if !claimed {
		if handled, err := s.moveWindow(x, y); handled || err != nil {
			return err
		}
	}

	frameW, frameH := s.frameSize()
	px, py := s.contentAt(x, y, frameW, frameH)

	if s.dev.on {
		if err := s.devPointer(x, y, px, py); err != nil {
			return err
		}
		return s.devTouch(frameW, frameH)
	}

	if err := s.hoverCursor(x, y, px, py); err != nil {
		return err
	}

	handled, err := s.menuPointer(x, y)
	if err != nil {
		return err
	}
	clicked, err := s.pointerMouse(x, y, px, py,
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft), handled)
	if err != nil {
		return err
	}
	return s.touch(clicked, frameW, frameH)
}

// contentPointAt maps the last cursor position into the painted page.
func (s *shell) contentPointAt(x, y int) (float64, float64) {
	frameW, frameH := s.frameSize()

	return s.contentAt(x, y, frameW, frameH)
}

type pointerDrag struct {
	active, mouse, moved bool
	x, y                 float64
	startX, startY       float64
}

func (s *shell) beginPointerDrag(x, y float64, mouse bool) (bool, error) {
	dragger, ok := s.app.(host.PointerDragger)
	if !ok {
		return false, nil
	}
	claimed, err := dragger.BeginDrag(s.ctx, x, y)
	if err == nil && claimed {
		s.gesture = pointerDrag{active: true, mouse: mouse, x: x, y: y, startX: x, startY: y}
		s.hold.cancel()
	}
	return claimed, err
}

func (s *shell) movePointerDrag(x, y float64) error {
	if x == s.gesture.x && y == s.gesture.y {
		return nil
	}
	if math.Abs(x-s.gesture.startX) > touchSlop || math.Abs(y-s.gesture.startY) > touchSlop {
		s.gesture.moved = true
	}
	s.gesture.x, s.gesture.y = x, y
	if dragger, ok := s.app.(host.PointerDragger); ok {
		return dragger.MoveDrag(s.ctx, x, y)
	}
	return nil
}

func (s *shell) endPointerDrag(tap bool) error {
	g := s.gesture
	s.gesture = pointerDrag{}
	dragger, ok := s.app.(host.PointerDragger)
	if !ok {
		return nil
	}
	if err := dragger.EndDrag(s.ctx); err != nil {
		return err
	}
	if tap {
		return s.tapAt(g.x, g.y)
	}
	return nil
}

// pressEdge reports the press this frame. A press and release that both land
// between two Update ticks leave only the release level, so a fast click
// during a long frame would otherwise reach the page as a release with no
// press. A release with no press this shell saw is therefore the remainder of
// a real click, and the release point is where to place it. A held button
// whose press was seen keeps wasDown true and never fires twice.
func pressEdge(down, released, wasDown bool) bool {
	if pressedNow(down, wasDown) {
		return true
	}

	return released && !down && !wasDown
}

// pressedNow reports the up-to-down edge of a pointer level.
func pressedNow(down, wasDown bool) bool {
	return down && !wasDown
}

func (s *shell) pointerMouse(
	x, y int, px, py float64, down, released, blocked bool,
) (bool, error) {
	pressed := pressEdge(down, released, s.mouseDown)
	handled := blocked || s.pointerScrollbar(x, y, down, pressed)
	if !handled && down && s.mouseDown && !s.gesture.active {
		s.dragScroll(y)
	}

	clicked := false
	if !handled {
		if pressed {
			clicked = true
			s.holdStart(px, py)
			if err := s.pressAt(px, py); err != nil {
				return false, err
			}
		} else if down {
			s.hold.move(px, py)
			if err := s.dragAt(px, py); err != nil {
				return false, err
			}
		}
	}
	s.mouseDown = down
	if released {
		if err := s.releaseAt(); err != nil {
			return clicked, err
		}
	}
	return clicked, nil
}

// touch tracks live fingers, scrolling, capture, and pinch.
// A system-generated mouse click suppresses the duplicate touch tap.
func (s *shell) touch(clicked bool, frameW, frameH int) error {
	ids := ebiten.AppendTouchIDs(nil)
	now := make([]touchPos, 0, len(ids))

	for _, id := range ids {
		x, y := ebiten.TouchPosition(id)
		y -= s.titlebarInset()
		if y < 0 {
			continue
		}
		now = append(now, touchPos{id: id, x: x, y: y})
	}

	return s.applyTouches(now, clicked, frameW, frameH)
}

// freshTouches returns the ids in now that are not in prev.
func freshTouches(now, prev []ebiten.TouchID) []ebiten.TouchID {
	var fresh []ebiten.TouchID

	for _, id := range now {
		if !hasTouch(prev, id) {
			fresh = append(fresh, id)
		}
	}

	return fresh
}

func hasTouch(ids []ebiten.TouchID, want ebiten.TouchID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}

	return false
}

const (
	clickGap  = 400 * time.Millisecond
	clickSlop = 4.0
)

// clickWatch counts rapid presses at one spot: one click, a double, or a
// triple.
type clickWatch struct {
	last  time.Time
	count int
	x, y  float64
}

// step records one press and returns the click count so far.
func (w *clickWatch) step(now time.Time, x, y float64) int {
	rapid := w.count > 0 && now.Sub(w.last) <= clickGap &&
		math.Abs(x-w.x) <= clickSlop && math.Abs(y-w.y) <= clickSlop

	if rapid {
		w.count++
	} else {
		w.count = 1
	}

	w.last, w.x, w.y = now, x, y

	return w.count
}

// cursorShape maps a page shape to an Ebiten cursor shape.
func cursorShape(shape host.Shape) ebiten.CursorShapeType {
	switch shape {
	case host.ShapeText:
		return ebiten.CursorShapeText
	case host.ShapePointer:
		return ebiten.CursorShapePointer
	case host.ShapeResizeEW:
		return ebiten.CursorShapeEWResize
	case host.ShapeResizeNS:
		return ebiten.CursorShapeNSResize
	default:
		return ebiten.CursorShapeDefault
	}
}

// cursorSupported reports that the platform has a mouse cursor to shape.
// wasm and mobile keep their own cursor.
func cursorSupported() bool {
	switch runtime.GOOS {
	case "js", "android", "ios":
		return false
	}

	return true
}

// applyCursor reads the hovered shape and sets the cursor only when it
// changed. The last shape stays on the shell for tests.
func (s *shell) applyCursor() {
	shaper, ok := s.app.(host.CursorShape)
	if !ok {
		return
	}

	s.setCursorShape(cursorShape(shaper.CursorShape()))
}

// applyThumbCursor shows the resize cursor over a scrollbar thumb. The page
// does not see the thumbs, so the shell names these shapes itself.
func (s *shell) applyThumbCursor(axis int) {
	switch axis {
	case dragVertical:
		s.setCursorShape(ebiten.CursorShapeNSResize)
	case dragHorizontal:
		s.setCursorShape(ebiten.CursorShapeEWResize)
	}
}

// setCursorShape records and applies one cursor shape, calling the test hook
// or Ebiten only when the shape changed.
func (s *shell) setCursorShape(next ebiten.CursorShapeType) {
	if next == s.cursor {
		return
	}

	s.cursor = next

	if s.setCursor != nil {
		s.setCursor(next)

		return
	}

	if cursorSupported() {
		ebiten.SetCursorShape(next)
	}
}

// hoverState remembers the page point of the last hover, the element the
// page painted after it, and the built generation it resolved against, so a
// repeat frame can skip it. The cursor fields remember the point,
// generation, and hover the page cursor was last applied for on the page
// path, so a still frame skips the cursor shape scan too.
type hoverState struct {
	x, y    float64
	seq     uint64
	id      string
	have    bool
	batched bool
	cx, cy  float64
	cgen    uint64
	cid     string
	cok     bool
}

// due reports whether the page hover needs a fresh pass. A point that moved
// always needs one. A still point on the same built generation reuses the
// last result. A new generation can move a box under the point, so that
// point resolves again too, unless the page paints a different element than
// the one this window pushed: then the page moved its hover itself, and
// pushing the window's stale point would relayout the page a second time for
// a hover it no longer shows. A screen that does not report its hover keeps
// resolving on every generation, as before.
func (h *hoverState) due(px, py float64, built uint64, pageHover string, reports bool) bool {
	if !h.have || h.x != px || h.y != py {
		return true
	}
	if h.seq == built {
		return false
	}

	return !reports || pageHover == h.id
}

// pageHover reports the element the page paints as hovered, when the screen
// can tell.
func (s *shell) pageHover() (string, bool) {
	hoverer, ok := s.app.(host.Hoverer)
	if !ok {
		return "", false
	}

	return hoverer.HoverID(), true
}

// hoverPushed records the hover the window just sent. The generation is the
// one the push left, not the one it resolved against: a hover that redrew
// the page must not read as a new generation on the next frame, or the same
// point would resolve again for a change it caused itself.
func (s *shell) hoverPushed(px, py float64) hoverState {
	id, _ := s.pageHover()

	return hoverState{x: px, y: py, seq: s.app.Generation(), id: id, have: true}
}

// hoverCursor sends page hover, then picks the cursor. Over a scrollbar
// thumb the resize shape pointerScrollbar applied stands, so the page
// shape never overwrites it. The first update runs before the first draw,
// so its hover push is deferred until one frame is committed: resolving it
// redraws the whole page for a cursor the user has not seen a frame under
// yet. Skipping the state record keeps it due, so the next update applies
// it after the first visible frame. A still point on the same generation
// with the same page hover keeps the applied page cursor: re-resolving it
// would walk the boxes and the link operations for the shape the window
// already shows. A thumb or drag path clears that record, so leaving the
// thumb re-applies the page shape.
func (s *shell) hoverCursor(x, y int, px, py float64) error {
	pageHover, reports := s.pageHover()
	pushed := false
	if s.hover.due(px, py, s.seq, pageHover, reports) && s.frameSeen() {
		if err := s.ptr().Hover(s.ctx, px, py); err != nil {
			return err
		}

		s.hover = s.hoverPushed(px, py)
		s.hover.batched = true
		pushed = true
		pageHover = s.hover.id
	}

	if s.dragAxis != dragNone {
		s.hover.cok = false
		s.applyThumbCursor(s.dragAxis)
		return nil
	}
	if axis, _, ok := s.scrollbarGrab(x, y); ok {
		s.hover.cok = false
		s.applyThumbCursor(axis)
		return nil
	}
	if !pushed && s.hover.cok && s.hover.cx == px && s.hover.cy == py &&
		s.hover.cgen == s.app.Generation() && s.hover.cid == pageHover {
		return nil
	}
	s.applyCursor()
	s.hover.cx, s.hover.cy = px, py
	s.hover.cgen = s.app.Generation()
	s.hover.cid = pageHover
	s.hover.cok = true

	return nil
}

// selector is the page range surface the window drives while the pointer is
// down. Deprecated: use host.RangeSelector. A page without it keeps the old
// click behavior.
type selector = host.RangeSelector

// pressAt sends a press, places the caret, then sends the click. The caret
// lands before the click handler runs, so focus and a selection that the
// handler sets are not blurred afterwards. A second press selects the word
// under the point and a third selects the line.
func (s *shell) pressAt(px, py float64) error {
	if claimed, err := s.beginPointerDrag(px, py, true); claimed || err != nil {
		return err
	}

	return host.Batch(s.ctx, s.app, func() error {
		if err := s.ptr().Press(s.ctx, px, py); err != nil {
			return err
		}

		sel, ok := s.app.(selector)
		if !ok {
			return s.ptr().Click(s.ctx, px, py)
		}

		if err := sel.SelectAt(s.ctx, px, py); err != nil {
			return err
		}

		if err := s.ptr().Click(s.ctx, px, py); err != nil {
			return err
		}

		s.dragActive = true
		s.dragX, s.dragY = px, py

		count := s.clicks.step(time.Now(), px, py)
		if count == doubleClickCount {
			return sel.SelectWordAt(s.ctx, px, py)
		}

		if count >= tripleClickCount {
			return sel.SelectLineAt(s.ctx, px, py)
		}

		return nil
	})
}

// dragAt extends the range while the button stays down and the point moves.
func (s *shell) dragAt(px, py float64) error {
	if s.gesture.active {
		return s.movePointerDrag(px, py)
	}
	if !s.dragActive || (px == s.dragX && py == s.dragY) {
		return nil
	}

	sel, ok := s.app.(selector)
	if !ok {
		return nil
	}

	s.dragX, s.dragY = px, py

	return sel.Drag(s.ctx, px, py)
}

// releaseAt ends a press and stops the drag.
func (s *shell) releaseAt() error {
	if s.gesture.active {
		return s.endPointerDrag(s.gesture.mouse && !s.gesture.moved)
	}
	s.dragActive = false
	s.hold.release()

	return s.ptr().Release(s.ctx)
}

// edgeScroll returns one scroll step when a drag point sits past the top or
// bottom edge of the viewport.
func edgeScroll(y, viewH, step int) int {
	if y < 0 {
		return -step
	}

	if y >= viewH {
		return step
	}

	return 0
}

// dragScroll scrolls while a selection drag sits past an edge. It reuses the
// wheel clamp, so the offset stops at the content ends.
func (s *shell) dragScroll(y int) {
	if !s.dragActive || s.stretched() || !s.allowPageScroll() {
		return
	}

	_, viewH := s.viewSize()

	step := edgeScroll(y, viewH, scrollStep)
	if step == 0 {
		return
	}

	_, contentH := s.contentSize()
	s.scrollY = clampAxis(s.scrollY+step, contentH-viewH)
}

// holdStart arms the held-press watch at a press point.
func (s *shell) holdStart(px, py float64) {
	s.hold.arm(time.Now(), px, py)
}

// holdFrame checks the held press from the frame loop, never a goroutine.
func (s *shell) holdFrame() error {
	return s.holdCheck(time.Now())
}

// holdCheck fires the screen's LongPress once when the press has been held
// in place long enough. A claim keeps the gesture for a selection drag.
func (s *shell) holdCheck(now time.Time) error {
	if !s.hold.due(now) {
		return nil
	}

	s.hold.fired = true

	lp, ok := s.app.(host.LongPresser)
	if !ok {
		return nil
	}

	claimed, err := lp.LongPress(s.ctx, s.hold.x, s.hold.y)
	if err != nil {
		return err
	}

	s.hold.claimed = claimed
	if claimed {
		s.dragActive = true
	}

	return nil
}

const (
	longPressDelay = 450 * time.Millisecond
	longPressSlop  = 8.0
)

// longPressWatch tracks one held press: it arms on the down edge, fires once
// when the delay passes without a move past the slop, and remembers whether
// the screen claimed the press.
type longPressWatch struct {
	start   time.Time
	x, y    float64
	armed   bool
	fired   bool
	claimed bool
}

// arm starts the hold at a press point.
func (w *longPressWatch) arm(now time.Time, x, y float64) {
	w.start = now
	w.x, w.y = x, y
	w.armed, w.fired, w.claimed = true, false, false
}

// move cancels an unclaimed hold that drifts past the slop. A claimed hold
// keeps the gesture, so the move can extend the selection.
func (w *longPressWatch) move(x, y float64) {
	if !w.armed || w.claimed {
		return
	}

	if math.Abs(x-w.x) > longPressSlop || math.Abs(y-w.y) > longPressSlop {
		w.armed = false
	}
}

// cancel drops a pending hold, such as when a second finger joins.
func (w *longPressWatch) cancel() {
	w.armed = false
	w.claimed = false
}

// release ends the press. claimed stays set, so the lift that ends a claimed
// hold is not delivered as a tap.
func (w *longPressWatch) release() {
	w.armed = false
}

// due reports the first frame past the delay. fired spends the hold, so
// LongPress runs once per press.
func (w *longPressWatch) due(now time.Time) bool {
	return w.armed && !w.fired && now.Sub(w.start) >= longPressDelay
}

// tapAt forwards a tap as press, caret, click, release. The caret lands
// before the click handler runs, matching pressAt, and a page without the
// selector keeps the old press, click, release behavior. The tap ends the
// press, so the held-press watch is disarmed first: a touch tap the window
// delivered itself must not fire a long press after the finger is up.
func (s *shell) tapAt(px, py float64) error {
	s.hold.release()

	return host.Batch(s.ctx, s.app, func() error {
		if err := s.ptr().Press(s.ctx, px, py); err != nil {
			return err
		}

		return s.finishTap(px, py)
	})
}

func (s *shell) finishTap(px, py float64) error {
	s.hold.release()
	if sel, ok := s.app.(host.Selector); ok {
		if err := sel.SelectAt(s.ctx, px, py); err != nil {
			return err
		}
	}

	if err := s.ptr().Click(s.ctx, px, py); err != nil {
		return err
	}

	return s.ptr().Release(s.ctx)
}
