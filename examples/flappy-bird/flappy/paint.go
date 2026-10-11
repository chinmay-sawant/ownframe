package flappy

import "github.com/chinmay-sawant/ownframe"

// paint moves the cached operations to match the game state. It runs every
// tick and never parses the HTML again.
func (a *App) paint() {
	d := a.page.Display()
	if d == nil {
		return
	}

	if a.page.Generation() != a.bound {
		a.bind(d)
	}

	a.paintScene(d)
	a.paintBird(d)
	a.paintPipes(d)
	a.paintText(d)
}

// setFill places one fill at r and shows or hides it.
func setFill(d *ownframe.Display, op *ownframe.DisplayOp, r rect, visible bool) {
	if op == nil {
		return
	}

	op.X = r.x * d.PointsPerPixel
	op.Y = r.y * d.PointsPerPixel
	op.W = r.w * d.PointsPerPixel
	op.H = r.h * d.PointsPerPixel

	if visible {
		op.Alpha = 1

		return
	}

	// The replay reads opacity, not Alpha, so a hidden fill collapses to
	// nothing and moves off the canvas: a zero-size rounded rect still
	// paints a corner sparkle at its place.
	op.X, op.Y = -1000, -1000
	op.W, op.H = 0, 0
	op.Alpha = 0
}
