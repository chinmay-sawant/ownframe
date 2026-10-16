package pet

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

type parts struct {
	eyeL        *ownframe.DisplayOp
	eyeR        *ownframe.DisplayOp
	mouth       *ownframe.DisplayOp
	heart       *ownframe.DisplayOp
	baseEyeL    rect
	baseEyeR    rect
	baseMouth   rect
	baseHeart   rect
	heartActive bool
	heartStart  time.Time
}

func (a *App) triggerHeart(now time.Time) {
	a.pm.heartActive = true
	a.pm.heartStart = now
}

func (a *App) bind(d *ownframe.Display) {
	bx := a.page.Boxes()
	a.pm.eyeL = fillAt(d, bx, "eye-left")
	a.pm.eyeR = fillAt(d, bx, "eye-right")
	a.pm.mouth = fillAt(d, bx, "pet-mouth")
	a.pm.heart = textAt(d, bx, "floating-heart")

	if a.pm.eyeL != nil {
		a.pm.baseEyeL = rect{x: a.pm.eyeL.X, y: a.pm.eyeL.Y, w: a.pm.eyeL.W, h: a.pm.eyeL.H}
	}
	if a.pm.eyeR != nil {
		a.pm.baseEyeR = rect{x: a.pm.eyeR.X, y: a.pm.eyeR.Y, w: a.pm.eyeR.W, h: a.pm.eyeR.H}
	}
	if a.pm.mouth != nil {
		a.pm.baseMouth = rect{x: a.pm.mouth.X, y: a.pm.mouth.Y, w: a.pm.mouth.W, h: a.pm.mouth.H}
	}
	if a.pm.heart != nil {
		a.pm.baseHeart = rect{x: a.pm.heart.X, y: a.pm.heart.Y, w: a.pm.heart.W, h: a.pm.heart.H}
		if !a.pm.heartActive {
			a.pm.heart.Alpha = 0.001
		}
	}
	a.bnd = a.page.Generation()
}

func (a *App) paint(now time.Time) {
	d := a.page.Display()
	if d == nil {
		return
	}
	if a.page.Generation() != a.bnd {
		a.bind(d)
	}
	a.paintEyes(now)
	a.paintMouth(now)
	a.paintHeart(now)
}
