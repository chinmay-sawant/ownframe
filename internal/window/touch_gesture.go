package window

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
