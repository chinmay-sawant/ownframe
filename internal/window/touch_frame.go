package window

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

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
			return s.app.Release(s.ctx)
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
