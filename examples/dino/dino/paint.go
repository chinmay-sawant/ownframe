package dino

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

	a.paintDino(d)
	a.paintObstacles(d)
	a.paintClouds(d)
	a.paintPebbles(d)
	a.setFill(d, a.parts.ground, rect{0, groundY - 2, sceneW, 2}, true)
	a.paintText()
}

// paintDino places the eight dinosaur fills for the current pose.
func (a *App) paintDino(d *ownframe.Display) {
	pose := a.game.dinoPose()

	for i := range dinoParts {
		a.setFill(d, a.parts.dino[i], pose[i], true)
	}
}

// setFill places one fill at r and shows or hides it. The view scales the
// scene and drops it down the page on a touch screen.
func (a *App) setFill(d *ownframe.Display, op *ownframe.DisplayOp, r rect, visible bool) {
	if op == nil {
		return
	}

	op.X = r.x * a.view.scale * d.PointsPerPixel
	op.Y = (r.y*a.view.scale + a.view.oy) * d.PointsPerPixel
	op.W = r.w * a.view.scale * d.PointsPerPixel
	op.H = r.h * a.view.scale * d.PointsPerPixel

	if visible {
		op.Alpha = 1

		return
	}

	// The replay reads opacity, not Alpha, so a hidden fill collapses to
	// nothing instead.
	op.W, op.H = 0, 0
	op.Alpha = 0
}
