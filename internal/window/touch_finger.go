package window

import "github.com/hajimehoshi/ebiten/v2"

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
