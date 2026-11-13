package flappy

import "github.com/chinmay-sawant/ownframe"

// paintScene drifts the clouds and the ground marks. The tick changes only
// the place along the sky or the ground; the layout holds the rest.
func (a *App) paintScene(d *ownframe.Display) {
	for i := range cloudMax {
		a.setX(d, a.parts.clouds[i], a.clouds[i])
	}

	for i := range stripeMax {
		a.setX(d, a.parts.stripes[i], a.stripes[i])
	}
}

// setX moves one operation along the ground. x is in game coordinates;
// the bound stage origin moves it into the page.
func (a *App) setX(d *ownframe.Display, op *ownframe.DisplayOp, x float64) {
	if op != nil {
		op.X = (x + a.ox) * d.PointsPerPixel
	}
}
