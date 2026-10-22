package window

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"math"
)

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
