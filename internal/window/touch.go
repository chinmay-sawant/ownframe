package window

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
	"time"
)

type touchCanceller interface{ TakeTouchCancel() bool }

func (s *shell) consumeTouchCancel(now []ebiten.TouchID) (bool, error) {
	request, ok := s.app.(touchCanceller)
	if ok && request.TakeTouchCancel() {
		s.touchCanceled = true
		s.fingers.fingers = nil
		s.fingers.span0 = 0
		s.fingers.multi = false
		s.hold.cancel()
		s.scrollMotion.cancel()
		if s.gesture.active && !s.gesture.mouse {
			if err := s.endPointerDrag(false); err != nil {
				return true, err
			}
		}
		if err := s.ptr().Release(s.ctx); err != nil {
			return true, err
		}
	}
	if !s.touchCanceled {
		return false, nil
	}
	if len(now) == 0 {
		s.touchCanceled = false
	}
	return true, nil
}

func (s *shell) routeTouchDrag(u touchUpdate, now []touchPos, w, h int) (bool, error) {
	if s.gesture.active && s.gesture.mouse {
		return false, nil
	}
	if !s.gesture.active && u.start != nil && len(now) == 1 {
		x, y := s.contentAt(u.start.x, u.start.y, w, h)
		if _, err := s.beginPointerDrag(x, y, false); err != nil {
			return false, err
		}
	}
	if !s.gesture.active {
		return false, nil
	}
	if len(now) == 0 {
		tap := u.tap != nil && !s.gesture.moved
		if err := s.endPointerDrag(false); err != nil {
			return true, err
		}
		return !tap, nil
	}
	if len(now) > 1 {
		return true, s.endPointerDrag(false)
	}
	x, y := s.contentAt(now[0].x, now[0].y, w, h)
	return true, s.movePointerDrag(x, y)
}

// pinchFingerCount is the fingers a pinch needs.
const pinchFingerCount = 2

// find returns the live finger with an id.
func (g *touchGesture) find(id ebiten.TouchID) (touchFinger, bool) {
	for _, f := range g.fingers {
		if f.id == id {
			return f, true
		}
	}

	return touchFinger{}, false
}

// stepPinch sets the zoom from the span of the first two fingers.
func (g *touchGesture) stepPinch() {
	if len(g.fingers) < pinchFingerCount {
		g.span0 = 0

		return
	}

	span := touchSpan(g.fingers[0], g.fingers[1])
	if g.span0 == 0 {
		g.span0 = span
		g.base = g.zoomOr1()
	}

	g.zoom = pinchScale(g.span0, span, g.base)
}

// zoomOr1 returns the pinch scale, or 1 before the first pinch.
func (g *touchGesture) zoomOr1() float64 {
	if g.zoom <= 0 {
		return 1
	}

	return g.zoom
}

func hasTouchPos(list []touchPos, id ebiten.TouchID) bool {
	for _, p := range list {
		if p.id == id {
			return true
		}
	}

	return false
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}

	return v
}

func (s *shell) applyTouches(now []touchPos, clicked bool, frameW, frameH int) error {
	lifted := len(s.fingers.fingers) > 0 && len(now) == 0
	u := s.fingers.frame(now, clicked)
	updated := time.Now()
	if len(now) > 0 || lifted {
		s.configureTouchScroll()
	}
	if u.start != nil {
		s.scrollMotion.begin(updated)
	} else if len(now) > 0 && u.dx == 0 && u.dy == 0 {
		s.scrollMotion.idle(updated)
	}
	if s.viewLocked() {
		s.holdView()
	} else if !s.touchZoomAllowed() {
		s.holdTouchZoom()
	}

	if handled, err := s.routeTouchDrag(u, now, frameW, frameH); handled || err != nil {
		return err
	}

	if err := s.touchPress(u, now, clicked, frameW, frameH); err != nil {
		return err
	}

	if err := s.touchMove(u, now, frameW, frameH); err != nil {
		return err
	}
	if lifted {
		s.scrollMotion.release()
	}

	if u.swipe != nil && !s.hold.claimed {
		if sw, ok := s.app.(host.Swiper); ok {
			if err := sw.Swipe(s.ctx, float64(u.swipe.dx), float64(u.swipe.dy)); err != nil {
				return err
			}
		}
	}

	if u.tap == nil {
		if lifted {
			s.hold.release()
			return s.ptr().Release(s.ctx)
		}
		return nil
	}

	if s.hold.claimed {
		return s.releaseAt()
	}

	tpx, tpy := s.contentAt(u.tap.x, u.tap.y, frameW, frameH)

	if s.hold.armed {
		return s.finishTap(tpx, tpy)
	}
	return s.tapAt(tpx, tpy)
}

// touchSlop is how far a finger may move before a tap becomes a drag.
const touchSlop = 8

// touchUpdate is one frame of touch movement. start is the first fresh
// finger of the frame, which arms the long press. swipe is a one-finger
// swipe that lifted past the tap slop.
type touchUpdate struct {
	dx, dy int
	tap    *touchPos
	start  *touchPos
	swipe  *swipeDelta
}

// touchGesture tracks fingers across frames: a moved finger drags the page,
// two fingers pinch, and a still finger that lifts taps.
type touchGesture struct {
	fingers []touchFinger
	zoom    float64
	span0   float64
	base    float64
	multi   bool
}

// frame advances the gesture one frame and returns what the shell should do.
func (g *touchGesture) frame(now []touchPos, swallowed bool) touchUpdate {
	var u touchUpdate
	g.multi = g.multi || len(now) >= pinchFingerCount

	next := make([]touchFinger, 0, len(now))
	for _, p := range now {
		f, ok := g.find(p.id)
		if !ok {
			next = append(next, touchFinger{
				id: p.id, x: p.x, y: p.y, startX: p.x, startY: p.y, eaten: swallowed,
			})

			if u.start == nil {
				u.start = &p
			}

			continue
		}

		dx := p.x - f.x
		dy := p.y - f.y
		if !f.moved && (absInt(p.x-f.startX) > touchSlop || absInt(p.y-f.startY) > touchSlop) {
			f.moved = true
		}

		if f.moved && len(now) == 1 {
			u.dx += dx
			u.dy += dy
		}

		f.x, f.y = p.x, p.y
		next = append(next, f)
	}

	for _, f := range g.fingers {
		if hasTouchPos(now, f.id) {
			continue
		}

		if !f.moved && !f.eaten && !g.multi {
			u.tap = &touchPos{id: f.id, x: f.x, y: f.y}
		} else if u.swipe == nil {
			u.swipe = liftSwipe(f, g.multi)
		}
	}

	g.fingers = next
	g.stepPinch()

	if len(next) == 0 {
		g.multi = false
		g.span0 = 0
	}

	return u
}

// touchMove applies one frame of finger movement. A claimed long press
// extends the selection and never scrolls. Any other move scrolls the page,
// clamped to the content ends.
func (s *shell) touchMove(u touchUpdate, now []touchPos, frameW, frameH int) error {
	if u.dx == 0 && u.dy == 0 {
		return nil
	}

	if s.hold.claimed && len(now) == 1 {
		px, py := s.contentAt(now[0].x, now[0].y, frameW, frameH)

		return s.dragAt(px, py)
	}

	s.hold.cancel()
	if err := s.ptr().Release(s.ctx); err != nil {
		return err
	}

	if s.stretched() || !s.allowPageScroll() {
		return nil
	}

	dx, dy := s.scrollMotion.move(u.dx, u.dy, time.Now())
	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	s.scrollX, s.scrollY = clampScroll(
		s.scrollX+dx, s.scrollY+dy, contentW, contentH, viewW, viewH,
	)

	return nil
}

func (s *shell) touchPress(u touchUpdate, now []touchPos, clicked bool, w, h int) error {
	if len(now) >= pinchFingerCount {
		armed := s.hold.armed || s.hold.claimed
		s.hold.cancel()
		s.scrollMotion.cancel()
		if armed {
			return s.ptr().Release(s.ctx)
		}
		return nil
	}
	if u.start == nil || clicked {
		return nil
	}
	x, y := s.contentAt(u.start.x, u.start.y, w, h)
	s.holdStart(x, y)
	return s.ptr().Press(s.ctx, x, y)
}

// touchScrollMotion bounds: sensitivity range, release keep factor, decay range.
const (
	touchSensMin     = .25
	touchSensMax     = 3
	touchReleaseKeep = .5
	touchDecayMin    = 1
	touchDecayMax    = 12
)

type touchScrollMotion struct {
	sensitivity, decay  float64
	vx, vy, rx, ry      float64
	lastTouch, lastStep time.Time
	dragging            bool
}

func newTouchScrollMotion(sensitivity, decay float64) touchScrollMotion {
	if sensitivity <= 0 || math.IsNaN(sensitivity) || math.IsInf(sensitivity, 0) {
		sensitivity = 1
	}
	if decay <= 0 || math.IsNaN(decay) || math.IsInf(decay, 0) {
		decay = 5
	}
	return touchScrollMotion{sensitivity: clampFloat(sensitivity, touchSensMin, touchSensMax), decay: clampFloat(decay, touchDecayMin, touchDecayMax)}
}

func (m *touchScrollMotion) defaults() {
	if m.sensitivity <= 0 {
		*m = newTouchScrollMotion(m.sensitivity, m.decay)
	}
}

func (m *touchScrollMotion) begin(now time.Time) {
	m.defaults()
	m.vx, m.vy, m.rx, m.ry = 0, 0, 0, 0
	m.lastTouch, m.lastStep, m.dragging = now, now, true
}

func (m *touchScrollMotion) configure(sensitivity, decay float64) {
	next := newTouchScrollMotion(sensitivity, decay)
	m.sensitivity, m.decay = next.sensitivity, next.decay
}

func (m *touchScrollMotion) release() {
	m.vx, m.vy = m.vx*touchReleaseKeep, m.vy*touchReleaseKeep
	m.dragging = false
}

func (m *touchScrollMotion) idle(now time.Time) {
	if m.dragging && now.Sub(m.lastTouch) > 45*time.Millisecond {
		m.vx, m.vy = 0, 0
	}
}

func (m *touchScrollMotion) move(dx, dy int, now time.Time) (int, int) {
	m.defaults()
	dt := now.Sub(m.lastTouch).Seconds()
	if dt > 0 && dt < .2 {
		vx, vy := -float64(dx)*m.sensitivity/dt, -float64(dy)*m.sensitivity/dt
		m.vx = scrollVelocity(m.vx, vx)
		m.vy = scrollVelocity(m.vy, vy)
	} else {
		m.vx, m.vy = 0, 0
	}
	m.lastTouch, m.lastStep = now, now
	x, rx := scrollPixels(-float64(dx)*m.sensitivity + m.rx)
	y, ry := scrollPixels(-float64(dy)*m.sensitivity + m.ry)
	m.rx, m.ry = rx, ry
	return x, y
}

func (m *touchScrollMotion) cancel() { m.vx, m.vy, m.dragging = 0, 0, false }

const scrollDtMax = .05 // seconds, max frame dt

const scrollVelCap = 5000 // px/s, velocity cap

const scrollStopSpeed = 24 // px/s, stop threshold

const scrollBlendOld = .65 // weight for old velocity

const scrollBlendNew = .35 // weight for new sample

const doubleClickCount = 2

const tripleClickCount = 3

func (m *touchScrollMotion) step(now time.Time) (int, int) {
	if m.dragging || m.sensitivity <= 0 {
		return 0, 0
	}
	dt := min(now.Sub(m.lastStep).Seconds(), scrollDtMax)
	m.lastStep = now
	if dt <= 0 {
		return 0, 0
	}
	keep := math.Exp(-m.decay * dt)
	x, rx := scrollPixels(m.vx*(1-keep)/m.decay + m.rx)
	y, ry := scrollPixels(m.vy*(1-keep)/m.decay + m.ry)
	m.rx, m.ry = rx, ry
	m.vx, m.vy = m.vx*keep, m.vy*keep
	if math.Abs(m.vx)+math.Abs(m.vy) < scrollStopSpeed {
		m.vx, m.vy = 0, 0
	}
	return x, y
}

func scrollVelocity(old, sample float64) float64 {
	if old == 0 {
		return clampFloat(sample, -scrollVelCap, scrollVelCap)
	}
	return clampFloat(old*scrollBlendOld+sample*scrollBlendNew, -scrollVelCap, scrollVelCap)
}

func scrollPixels(v float64) (int, float64) {
	pixels := int(v)
	return pixels, v - float64(pixels)
}

func clampFloat(v, low, high float64) float64 { return max(low, min(v, high)) }

func (s *shell) stepTouchScroll(now time.Time) {
	if s.scrollMotion.dragging || s.scrollMotion.vx == 0 && s.scrollMotion.vy == 0 {
		return
	}
	s.configureTouchScroll()
	if s.stretched() || !s.allowPageScroll() {
		s.scrollMotion.cancel()
		return
	}
	dx, dy := s.scrollMotion.step(now)
	if dx == 0 && dy == 0 {
		return
	}
	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	x, y := clampScroll(s.scrollX+dx, s.scrollY+dy, contentW, contentH, viewW, viewH)
	if x != s.scrollX+dx || y != s.scrollY+dy {
		s.scrollMotion.cancel()
	}
	s.scrollX, s.scrollY = x, y
}

func (s *shell) configureTouchScroll() {
	if config, ok := s.app.(host.TouchScrollConfig); ok {
		s.scrollMotion.configure(config.TouchScrollSensitivity(), config.TouchScrollDeceleration())
	}
}

// swipeMin is how far a one-finger drag must travel to count as a swipe.
const swipeMin = 24

// swipeDelta is a one-finger swipe that lifted past the tap slop.
type swipeDelta struct {
	dx, dy int
}

// liftSwipe returns the swipe a lifting finger made, or nil. A finger that
// never moved is a tap, an eaten finger belongs to the screen, and a
// gesture with a second finger never swipes.
func liftSwipe(f touchFinger, multi bool) *swipeDelta {
	if !f.moved || f.eaten || multi {
		return nil
	}

	dx := f.x - f.startX
	dy := f.y - f.startY
	if absInt(dx) < swipeMin && absInt(dy) < swipeMin {
		return nil
	}

	return &swipeDelta{dx: dx, dy: dy}
}

// touchPos is one finger position this frame.
type touchPos struct {
	id   ebiten.TouchID
	x, y int
}

// touchFinger is one finger in flight.
type touchFinger struct {
	id     ebiten.TouchID
	x, y   int
	startX int
	startY int
	moved  bool
	eaten  bool
}

const (
	minZoom = 0.25
	maxZoom = 4
)

// pinchScale returns the zoom for a pinch: the base zoom times the change in
// finger span, clamped to a usable range.
func pinchScale(start, now, base float64) float64 {
	if start <= 0 || base <= 0 {
		return 1
	}

	z := base * now / start
	if z < minZoom {
		return minZoom
	}

	if z > maxZoom {
		return maxZoom
	}

	return z
}

// touchSpan is the distance between two fingers.
func touchSpan(a, b touchFinger) float64 {
	dx := float64(a.x - b.x)
	dy := float64(a.y - b.y)

	return math.Hypot(dx, dy)
}
