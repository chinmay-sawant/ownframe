package window

import "github.com/hajimehoshi/ebiten/v2"

type windowDrag struct {
	active, moved, down bool
	x, y                int
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

func (s *shell) moveWindow(x, y int) (bool, error) {
	if s.draggable == nil {
		return false, nil
	}
	handled, move, click, dx, dy := s.windowDrag.step(x, y,
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft), s.draggable(x, y))
	if move {
		// Dragging a maximized window restores it first, in the style of
		// VS Code: a fullscreen-sized frame cannot move anywhere useful,
		// so the drag picks up the windowed rect instead of sliding the
		// workarea around. WSL manual maximize is plain client geometry,
		// so restore the saved size at the live position and keep the
		// grab offset; a WM maximize restores and lets the next frames
		// carry the drag once the configure lands.
		if sw, sh, restored := restoreManualForDrag(); restored {
			if validPlacementRect(sw, sh) {
				ebiten.SetWindowSize(sw, sh)
			}
		} else if ebiten.IsWindowMaximized() {
			ebiten.RestoreWindow()

			return true, nil
		}
		wx, wy := ebiten.WindowPosition()
		ebiten.SetWindowPosition(wx+dx, wy+dy)
	}
	if click {
		px, py := s.contentPointAt(s.windowDrag.x, s.windowDrag.y)
		return true, s.ptr().Click(s.ctx, px, py)
	}
	return handled, nil
}
