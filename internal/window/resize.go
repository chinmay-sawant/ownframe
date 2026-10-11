package window

import (
	"errors"
	"time"
)

// relayoutEvery caps how often a drag relayouts the page. A size change
// inside that window keeps the previous frame scaled to the window instead.
const relayoutEvery = 100 * time.Millisecond

// resize commits a pending window size to the page. While the size keeps
// changing it relayouts at most every relayoutEvery; a size that held still
// for one update commits at once, so a drag ends on its final size.
func (s *shell) resize() error {
	moving := s.moving
	s.moving = false

	wantW, wantH := s.app.Clamp(s.pendingW, s.pendingH)
	haveW, haveH := s.app.Size()
	if wantW == haveW && wantH == haveH {
		return nil
	}

	if moving && time.Since(s.lastRelayout) < relayoutEvery {
		s.skipped++

		return nil
	}

	s.app.SetSize(wantW, wantH)

	if err := s.app.Redraw(s.ctx); err != nil {
		return err
	}

	s.lastRelayout = time.Now()
	s.commits++

	if err := s.syncImage(); err != nil && !errors.Is(err, errNoImage) {
		return err
	}

	return s.refreshState()
}

// Layout records the outside size. The commit happens in resize on the
// next update, so a burst of events between updates costs one relayout.
func (s *shell) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth < 1 {
		outsideWidth = 1
	}

	if outsideHeight < 1 {
		outsideHeight = 1
	}

	if outsideWidth != s.pendingW || outsideHeight != s.pendingH {
		s.imeLayoutChanged(outsideWidth, outsideHeight)
		s.moving = true
	}

	s.pendingW = outsideWidth
	s.pendingH = outsideHeight
	s.screenW = outsideWidth
	s.screenH = outsideHeight

	return outsideWidth, outsideHeight
}
