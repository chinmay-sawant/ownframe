package window

import (
	"errors"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// relayoutEvery caps how often a drag relayouts the page. A size change
// inside that window keeps the previous frame scaled to the window instead.
const relayoutEvery = 100 * time.Millisecond

// resize commits the page viewport to the page: the pending window size
// minus the devtools dock when it is open. While the size keeps changing it
// relayouts at most every relayoutEvery; a size that held still for one
// update commits at once, so a drag ends on its final size.
func (s *shell) resize() error {
	moving := s.moving
	s.moving = false

	completeManualMaximize(s.pendingW, s.pendingH)

	if pendingStartMaximize {
		// A remembered WM maximize from placement: the pre-loop call never
		// takes effect, so maximize once the window exists.
		pendingStartMaximize = false
		ebiten.MaximizeWindow()
	}

	wantW, wantH := s.viewOf(s.pendingW, s.pendingH)
	wantW, wantH = s.app.Clamp(wantW, wantH)
	haveW, haveH := s.app.Size()
	if wantW == haveW && wantH == haveH {
		s.syncPlacement()

		return nil
	}

	if s.windowDrag.active {
		// A titlebar drag moves the window; it does not resize it. A
		// size report that lands mid-drag (monitor scale seam) must not
		// trigger a full relayout storm under the hand. The commit lands
		// on release, when the drag goes inactive and the sizes still
		// differ. The in-between frames keep the previous frame scaled.
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

	// maximizeNudge re-asserts the configured window size while maximized.
	// A client-set size travels the normal presentation path, so on systems
	// whose compositor mispresents a maximized undecorated window the
	// re-assert can snap it back. Where the size already matches it is a
	// no-op: same-size sets change nothing. Position is left alone; only
	// the size is ever re-asserted.
	if ebiten.IsWindowMaximized() {
		ebiten.SetWindowSize(s.pendingW, s.pendingH)
	}

	s.syncPlacement()

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
