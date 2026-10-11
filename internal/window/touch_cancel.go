package window

import "github.com/hajimehoshi/ebiten/v2"

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
		if err := s.app.Release(s.ctx); err != nil {
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
