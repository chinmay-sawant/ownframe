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

func (s *shell) moveWindow(x, y int) (bool, error) {
	if s.draggable == nil {
		return false, nil
	}
	handled, move, click, dx, dy := s.windowDrag.step(x, y,
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft), s.draggable(x, y))
	if move {
		wx, wy := ebiten.WindowPosition()
		ebiten.SetWindowPosition(wx+dx, wy+dy)
	}
	if click {
		px, py := s.contentPointAt(s.windowDrag.x, s.windowDrag.y)
		return true, s.ptr().Click(s.ctx, px, py)
	}
	return handled, nil
}
