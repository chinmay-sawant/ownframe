package pet

import (
	"math"
	"time"
)

func (a *App) paintEyes(now time.Time) {
	if a.pm.eyeL == nil || a.pm.eyeR == nil {
		return
	}
	open := 1.0
	if a.state.Sleeping {
		open = 0.15
	} else {
		sec := now.Sub(time.Unix(0, 0)).Seconds()
		cycle := math.Mod(sec, 3.2)
		if cycle < 0.16 {
			progress := cycle / 0.16
			open = clamp(1.0-math.Sin(progress*math.Pi), 0.15, 1.0)
		}
	}

	hL := a.pm.baseEyeL.h * open
	a.pm.eyeL.H = hL
	a.pm.eyeL.Y = a.pm.baseEyeL.y + (a.pm.baseEyeL.h-hL)/2

	hR := a.pm.baseEyeR.h * open
	a.pm.eyeR.H = hR
	a.pm.eyeR.Y = a.pm.baseEyeR.y + (a.pm.baseEyeR.h-hR)/2
}

func (a *App) paintMouth(now time.Time) {
	if a.pm.mouth == nil {
		return
	}
	if a.state.Sleeping {
		a.pm.mouth.W = a.pm.baseMouth.w * 0.8
		a.pm.mouth.X = a.pm.baseMouth.x + (a.pm.baseMouth.w-a.pm.mouth.W)/2
		return
	}

	sec := now.Sub(time.Unix(0, 0)).Seconds()
	rate := 2.5
	if now.Before(a.state.PlayUntil) {
		rate = 12.0
	} else if now.Before(a.state.PettingUntil) {
		rate = 7.0
	}

	osc := math.Sin(sec * rate)
	w := a.pm.baseMouth.w * (1.0 + 0.2*osc)
	a.pm.mouth.W = w
	a.pm.mouth.X = a.pm.baseMouth.x - (w-a.pm.baseMouth.w)/2
}

func (a *App) paintHeart(now time.Time) {
	if a.pm.heart == nil {
		return
	}
	if !a.pm.heartActive {
		a.pm.heart.Alpha = 0.001
		return
	}

	dur := now.Sub(a.pm.heartStart).Seconds()
	const maxDur = 1.6
	if dur >= maxDur {
		a.pm.heartActive = false
		a.pm.heart.Alpha = 0.001
		a.pm.heart.Y = a.pm.baseHeart.y
		return
	}

	p := dur / maxDur
	a.pm.heart.Y = a.pm.baseHeart.y - 35.0*p
	a.pm.heart.Alpha = clamp(1.0-p, 0.001, 1.0)
}
