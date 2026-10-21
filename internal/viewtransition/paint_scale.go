package viewtransition

import "github.com/chinmay-sawant/ownframe"

func scaleOp(op *ownframe.DisplayOp, ppp float64, box Rect, scale float64) {
	if ppp <= 0 {
		ppp = cssDPI / pointsPerInch
	}
	cx, cy := (box.X+box.W/halfDivisor)/ppp, (box.Y+box.H/halfDivisor)/ppp
	op.X = cx + (op.X-cx)*scale
	op.Y = cy + (op.Y-cy)*scale
	op.W *= scale
	op.H *= scale
	op.Size *= scale
	op.Width *= scale
}
