package flappy

import "github.com/chinmay-sawant/ownframe"

// paintScene drifts the clouds and the ground marks. The tick changes only
// the place along the sky or the ground; the layout holds the rest.
func (a *App) paintScene(d *ownframe.Display) {
	for i := range cloudMax {
		setX(a.parts.clouds[i], a.clouds[i]*d.PointsPerPixel)
	}

	for i := range stripeMax {
		setX(a.parts.stripes[i], a.stripes[i]*d.PointsPerPixel)
	}
}

// setX moves one operation along the ground.
func setX(op *ownframe.DisplayOp, x float64) {
	if op != nil {
		op.X = x
	}
}
