package window

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// dragMinVisibleX keeps at least this many pixels of window width
	// inside the outer monitor bound, so a titlebar can be grabbed back.
	dragMinVisibleX = 160
	// dragMinVisibleY keeps at least this many pixels of window height
	// inside the outer monitor bound.
	dragMinVisibleY = 32
	// dragMinApplyPx drops micro-jitter moves. Deltas below this never
	// reach the compositor; the pending remainder re-accumulates them, so
	// slow drags stay exact while the request rate collapses.
	dragMinApplyPx = 2
	// dragCursorJumpPx is the raw cursor step that cannot be a hand.
	// Captured drags stay under about 400px of smoothed motion, and a
	// fast real sample still lands under this. A step of about one
	// monitor width is WSLg changing the cursor's coordinate space for a
	// frame or two, often after the monitor name has already settled.
	// Applying it throws the window, and the low-pass then smears that
	// throw across the following frames. One such sample is ignored and
	// the smoother rewound. A second sample that agrees with it adopts
	// the new coordinates without moving, so a stuck rebase keeps tracking.
	dragCursorJumpPx = 720
	// dragApplySlack is how far a presented position may sit from the
	// position we requested, or from the previous sample, and still be
	// that request. A monitor-origin rebase is hundreds of pixels away
	// from both.
	dragApplySlack = 48
	// smoothDivisor is the low-pass strength on the drag cursor. Each
	// frame the smoothed cursor moves 1/smoothDivisor of the way to the
	// raw sample, so single-frame spikes fade geometrically while a real
	// hand trails by about one frame. Snap on press keeps clicks exact.
	smoothDivisor = 2
)

type windowDrag struct {
	active, moved, down bool
	x, y                int
	// fx, fy is the low-passed cursor. The raw sample over WSLg jumps
	// frame to frame; smoothing it removes the scuff while the hand
	// trails by about one frame. Snapped to raw on press.
	fx, fy float64
	// Gesture tracking recovers the physical hand motion from the
	// window-relative cursor: lsx,lsy is the last smoothed cursor,
	// lwx,lwy the last window position, and pendx,pendy the unapplied
	// remainder. A window move shifts the next cursor reading by its own
	// mirror, so the raw delta alone would feed the frame's motion back
	// into itself and oscillate. Adding the window shift back cancels our
	// own motion; a steady hand then reads exactly zero and the window
	// holds instead of sailing off on a stale offset.
	lsx, lsy, lwx, lwy int
	pendx, pendy       int
	tracked            bool
	// lrx, lry is the last raw cursor sample. When the raw hand is
	// frozen, any recovered motion comes from the presented window
	// position wobbling under WSLg, not from the user. Draining only
	// the smoothing tail then holds the frame dead still instead of
	// integrating present noise into a permanent creep.
	lrx, lry int
	// bad is set when the last raw sample was a cursor jump we refused
	// to apply. badX, badY is that sample. A second jump that lands near
	// it is a stable coordinate change and is adopted. A return to the
	// previous anchor is a one-frame glitch and is tracked normally.
	bad        bool
	badX, badY int
	// reqx, reqy is the last position handed to the compositor. haveReq
	// is false until the first request of this gesture. A presented
	// position that matches neither this nor the previous sample is a
	// monitor-origin rebase, not the hand.
	reqx, reqy int
	haveReq    bool
	// jumpDrop, jumpSnap, and rebased describe the latest prepareDrag
	// call so the trace can say which guard fired.
	jumpDrop bool
	jumpSnap bool
	rebased  bool
	rebaseX  int
	rebaseY  int
}

// step uses window-local coordinates; movement keeps the grab point fixed.
func (d *windowDrag) step(x, y int, down, allowed bool) (handled, move, click bool, dx, dy int) {
	pressed := down && !d.down
	d.down = down
	if !d.active {
		if !pressed || !allowed {
			return
		}
		d.active = true
		d.moved = false
		d.x, d.y = x, y
		d.tracked = false
	}
	handled = true
	if !down {
		click = !d.moved
		d.active = false
		return
	}
	dx, dy = x-d.x, y-d.y
	if !d.moved && dx*dx+dy*dy < 16 {
		return
	}
	d.moved = true
	move = dx != 0 || dy != 0
	return
}

// recoverMotion returns the physical hand motion for one frame: the
// smoothed cursor shift plus the window shift. The cursor is reported
// window-relative, so a window move shifts the next reading by its own
// mirror; adding the window shift back cancels our own motion instead of
// feeding it into the next delta.
func recoverMotion(sx, sy, lsx, lsy, wx, wy, lwx, lwy int) (dx, dy int) {
	return (sx - lsx) + (wx - lwx), (sy - lsy) + (wy - lwy)
}

// restoreManualForDrag exits a manual maximize when a titlebar drag starts,
// so a workarea-sized frame can move as a windowed one. It returns the
// windowed size to install, clamped to the live monitor.
func restoreManualForDrag() (int, int, bool) {
	if !manualMax.active {
		return 0, 0, false
	}

	sw, sh := manualMax.sw, manualMax.sh
	manualMax.active = false

	if !validPlacementRect(sw, sh) {
		return 0, 0, true
	}

	if monitor := ebiten.Monitor(); monitor != nil {
		if monW, monH := monitor.Size(); monW > 0 && monH > 0 {
			sw = min(sw, monW)
			sh = min(sh, monH)
		}
	}

	return sw, sh, true
}

// dragOuterBound is a generous outer bound for window positions: the summed
// monitor widths and heights. Ebiten reports sizes but not origins, so this
// cannot trace the true union; it only bites in extreme positions where any
// layout would have lost the window. Zero when no monitor reports a size.
func dragOuterBound() (int, int) {
	var w, h int

	for _, m := range ebiten.AppendMonitors(nil) {
		mw, mh := m.Size()
		if mw > 0 && mh > 0 {
			w += mw
			h += mh
		}
	}

	return w, h
}

// negligibleDrag reports whether a drag move is micro-jitter worth
// withholding from the compositor. The pending remainder keeps the debt,
// so withholding only delays, never loses.
func negligibleDrag(dx, dy int) bool {
	return max(absDrag(dx), absDrag(dy)) < dragMinApplyPx
}

func absDrag(v int) int {
	if v < 0 {
		return -v
	}

	return v
}

// smoothCursor moves prev 1/smoothDivisor of the way to the raw sample, so
// single-frame spikes fade geometrically while a real hand is tracked.
func smoothCursor(prev float64, raw int) float64 {
	return prev + (float64(raw)-prev)/smoothDivisor
}

// clampDragTarget keeps part of a ww by wh window inside a bw by bh bound
// so a flung drag cannot strand the frame where no titlebar is grabbable.
// Unknown bounds and oversized windows skip the clamp on that axis.
func clampDragTarget(wx, wy, ww, wh, bw, bh int) (int, int) {
	if bw > 0 && ww > 0 && ww-dragMinVisibleX < bw {
		low, high := -(ww - dragMinVisibleX), bw-dragMinVisibleX
		if low > high {
			low = high
		}

		wx = min(max(wx, low), high)
	}

	if bh > 0 && wh > 0 && wh-dragMinVisibleY < bh {
		low, high := -(wh - dragMinVisibleY), bh-dragMinVisibleY
		if low > high {
			low = high
		}

		wy = min(max(wy, low), high)
	}

	return wx, wy
}

// dragSample is one motion frame. raw is the cursor Ebiten just reported,
// sx/sy is the low-passed cursor, and wx/wy is the presented window position.
type dragSample struct {
	rawX, rawY int
	sx, sy     int
	wx, wy     int
}

// windowRebased reports that the presented position is neither the previous
// sample nor the position we requested, including a partial step toward
// that request. That jump is a monitor-origin change. Counting it as hand
// motion slides the window off the cursor.
func windowRebased(prevX, prevY, reqX, reqY, obsX, obsY int) bool {
	return !dragAxisExplained(prevX, reqX, obsX) || !dragAxisExplained(prevY, reqY, obsY)
}

func dragAxisExplained(prev, req, obs int) bool {
	if absDrag(obs-req) <= dragApplySlack || absDrag(obs-prev) <= dragApplySlack {
		return true
	}

	return betweenDrag(prev, obs, req)
}

func betweenDrag(a, x, b int) bool {
	if a > b {
		a, b = b, a
	}

	return x >= a && x <= b
}

func cursorJump(dx, dy int) bool {
	return absDrag(dx) >= dragCursorJumpPx || absDrag(dy) >= dragCursorJumpPx
}

// refuseCursorJump drops a raw cursor step of dragCursorJumpPx or more.
// The smoother has already moved halfway toward that sample, so the
// filtered cursor is rewound to the last trusted point; otherwise the
// next frame applies the tail. The raw anchor stays put. A glitch that
// returns on the next sample is then a small real delta. Two refused
// samples that agree are a coordinate space that stuck: the anchors snap
// to it and the jump itself is not applied, so the window stays under
// the hand. haveReq is left set so a position already in flight can
// still cancel through the cursor mirror.
func (d *windowDrag) refuseCursorJump(s dragSample) bool {
	if !cursorJump(s.rawX-d.lrx, s.rawY-d.lry) {
		d.bad = false

		return false
	}

	if d.bad && !cursorJump(s.rawX-d.badX, s.rawY-d.badY) {
		d.jumpSnap = true
		d.fx, d.fy = float64(s.rawX), float64(s.rawY)
		d.lsx, d.lsy = s.rawX, s.rawY
		d.lrx, d.lry = s.rawX, s.rawY
		d.lwx, d.lwy = s.wx, s.wy
		d.pendx, d.pendy = 0, 0
		d.bad = false

		return true
	}

	d.jumpDrop = true
	d.bad = true
	d.badX, d.badY = s.rawX, s.rawY
	d.fx, d.fy = float64(d.lsx), float64(d.lsy)

	return true
}

// prepareDrag updates the gesture anchors and returns the delta to add to
// the presented window position. ok is false when the sample stays off the
// compositor: the gesture just anchored, the delta is still micro-jitter,
// or the raw cursor jumped and must not move the window.
func (d *windowDrag) prepareDrag(s dragSample) (dx, dy int, ok bool) {
	d.jumpDrop = false
	d.jumpSnap = false
	d.rebased = false
	if !d.tracked {
		d.lsx, d.lsy = s.sx, s.sy
		d.lwx, d.lwy = s.wx, s.wy
		d.lrx, d.lry = s.rawX, s.rawY
		d.pendx, d.pendy = 0, 0
		d.tracked = true
		d.haveReq = false
		d.bad = false

		return 0, 0, false
	}

	if d.haveReq && windowRebased(d.lwx, d.lwy, d.reqx, d.reqy, s.wx, s.wy) {
		d.rebased = true
		d.rebaseX, d.rebaseY = d.lwx, d.lwy
		d.lwx, d.lwy = s.wx, s.wy
		d.haveReq = false
	}

	if d.refuseCursorJump(s) {
		return 0, 0, false
	}

	var mx, my int
	if s.rawX == d.lrx && s.rawY == d.lry {
		// Hand frozen: ignore the window shift and drain only the
		// smoothing tail. The shift is present wobble, and feeding
		// it back is what kept the window creeping on its own.
		mx, my = s.sx-d.lsx, s.sy-d.lsy
	} else {
		mx, my = recoverMotion(s.sx, s.sy, d.lsx, d.lsy, s.wx, s.wy, d.lwx, d.lwy)
	}

	d.lsx, d.lsy = s.sx, s.sy
	d.lwx, d.lwy = s.wx, s.wy
	d.lrx, d.lry = s.rawX, s.rawY
	d.pendx += mx
	d.pendy += my
	if negligibleDrag(d.pendx, d.pendy) {
		return d.pendx, d.pendy, false
	}

	return d.pendx, d.pendy, true
}

// commitDrag records the position the compositor was given and drops any
// part of the delta the outer bound clamped away, so the bound cannot
// build a debt that jumps back on the next frame.
func (d *windowDrag) commitDrag(wx, wy, nx, ny int, clamped bool) {
	d.pendx -= nx - wx
	d.pendy -= ny - wy
	if clamped {
		d.pendx, d.pendy = 0, 0
	}

	d.reqx, d.reqy = nx, ny
	d.haveReq = true
}

func (s *shell) moveWindow(x, y int) (bool, error) {
	if s.draggable == nil {
		return false, nil
	}
	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if down && !s.windowDrag.down {
		// Fresh press: snap the filtered cursor so clicks stay exact.
		s.windowDrag.fx, s.windowDrag.fy = float64(x), float64(y)
	}
	if down {
		s.windowDrag.fx = smoothCursor(s.windowDrag.fx, x)
		s.windowDrag.fy = smoothCursor(s.windowDrag.fy, y)
	}
	sx, sy := int(math.Round(s.windowDrag.fx)), int(math.Round(s.windowDrag.fy))
	allowed := s.draggable(x, y) && !s.devOwnsPoint(x, y)
	handled, move, click, _, _ := s.windowDrag.step(sx, sy,
		down, allowed)
	if !down {
		s.windowDrag.tracked = false
	}
	if move {
		if manualMax.learning || manualMax.restoring {
			// A workarea probe is in flight; keep the anchor fresh and
			// touch nothing so the drag cannot kill the dance.
			s.windowDrag.x, s.windowDrag.y = sx, sy

			return true, nil
		}
		// Dragging a maximized window restores it first, in the style of
		// VS Code: a fullscreen-sized frame cannot move anywhere useful,
		// so the drag picks up the windowed rect instead of sliding the
		// workarea around. WSL manual maximize is plain client geometry,
		// so restore the saved size at the live position and keep the
		// grab offset; a WM maximize restores and lets the next frames
		// carry the drag once the configure lands.
		winW, winH := ebiten.WindowSize()
		if sw, sh, restored := restoreManualForDrag(); restored {
			if validPlacementRect(sw, sh) {
				ebiten.SetWindowSize(sw, sh)
				winW, winH = sw, sh
			}
		} else if ebiten.IsWindowMaximized() {
			ebiten.RestoreWindow()

			return true, nil
		}
		wx, wy := ebiten.WindowPosition()
		mon := currentMonitorName()
		dx, dy, ok := s.windowDrag.prepareDrag(dragSample{
			rawX: x,
			rawY: y,
			sx:   sx,
			sy:   sy,
			wx:   wx,
			wy:   wy,
		})
		if s.dragTrace && s.windowDrag.rebased {
			_, _ = fmt.Fprintf(os.Stderr, "[drag] rebase %d,%d -> %d,%d mon=%q\n",
				s.windowDrag.rebaseX, s.windowDrag.rebaseY, wx, wy, mon)
		}
		if !ok {
			// Below the apply threshold, or a cursor coordinate jump.
			// The pending remainder accumulates across tiny samples, so
			// a slow drag stays exact. A refused jump leaves that debt
			// alone; adopting a stable coordinate clears it.
			switch {
			case s.dragTrace && s.windowDrag.jumpSnap:
				_, _ = fmt.Fprintf(os.Stderr, "[drag] jump snap raw=%d,%d win=%d,%d mon=%q\n",
					x, y, wx, wy, mon)
			case s.dragTrace && s.windowDrag.jumpDrop:
				_, _ = fmt.Fprintf(os.Stderr, "[drag] jump raw=%d,%d win=%d,%d mon=%q\n",
					x, y, wx, wy, mon)
			case s.dragTrace:
				_, _ = fmt.Fprintf(os.Stderr, "[drag] tiny pend=%d,%d win=%d,%d mon=%q\n",
					s.windowDrag.pendx, s.windowDrag.pendy, wx, wy, mon)
			}

			return true, nil
		}
		nx, ny := wx+dx, wy+dy
		clamped := false
		if bw, bh := dragOuterBound(); bw > 0 && bh > 0 {
			cx, cy := clampDragTarget(nx, ny, winW, winH, bw, bh)
			clamped = cx != nx || cy != ny
			nx, ny = cx, cy
		}
		reqAt := time.Now()
		ebiten.SetWindowPosition(nx, ny)
		reqTook := time.Since(reqAt)
		if s.dragTrace {
			_, _ = fmt.Fprintf(os.Stderr, "[drag] cursor=%d,%d phys=%d,%d win=%d,%d size=%dx%d target=%d,%d clamped=%v max=%v mon=%q took=%dus\n",
				sx, sy, dx, dy, wx, wy, winW, winH, nx, ny, clamped, manualMax.active, mon, reqTook.Microseconds())
		}
		s.windowDrag.commitDrag(wx, wy, nx, ny, clamped)
	}
	if click {
		px, py := s.contentPointAt(s.windowDrag.x, s.windowDrag.y)
		return true, s.ptr().Click(s.ctx, px, py)
	}
	return handled, nil
}
