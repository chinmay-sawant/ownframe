package flappy

import "github.com/chinmay-sawant/ownframe"

// paint moves the cached operations to match the game state. It runs every
// tick and never parses the HTML again.
func (a *App) paint() {
	d := a.page.Probe().Display()
	if d == nil {
		return
	}

	if a.page.Probe().Generation() != a.bound {
		a.bind(d)
	}

	a.paintScene(d)
	a.paintBird(d)
	a.paintPipes(d)
	a.paintText(d)
}

// setFill places one fill at r and shows or hides it. r is in game
// coordinates from the stage corner; the bound stage origin moves it into
// the page so a centered stage paints under its own boxes.
func (a *App) setFill(d *ownframe.Display, op *ownframe.DisplayOp, r rect, visible bool) {
	if op == nil {
		return
	}

	op.X = (r.x + a.ox) * d.PointsPerPixel
	op.Y = (r.y + a.oy) * d.PointsPerPixel
	op.W = r.w * d.PointsPerPixel
	op.H = r.h * d.PointsPerPixel

	if visible {
		op.Alpha = 1

		return
	}

	// The replay reads opacity, not Alpha, so a hidden fill collapses to
	// nothing and moves off the canvas: a zero-size rounded rect still
	// paints a corner sparkle at its place.
	op.X, op.Y = offscreen, offscreen
	op.W, op.H = 0, 0
	op.Alpha = 0
}
