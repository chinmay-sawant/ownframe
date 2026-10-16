package eyes

import "math"

func animatePose(p eyePose, age float64) (eyeStyle, eyeStyle) {
	l, r := p.left, p.right
	wave := math.Sin(age * 2 * math.Pi / 3)
	switch p.motion {
	case motionRelaxed:
		l.open += .07 * wave
		r.open += .07 * wave
	case motionHappy:
		bright := math.Max(0, math.Sin(age*2.1))
		grin := math.Max(0, math.Sin(age*2.1-1.3))
		l.open += .15*bright - .2*grin
		l.top -= .08 * bright
		l.bottom += .16 * grin
		l.gazeY -= 2 * bright
		r.bottom = l.bottom
		r.open, r.top, r.gazeY = l.open, l.top, l.gazeY
	case motionWorried:
		wobble := 1.4 * math.Sin(age*2.1)
		l.slant += wobble
		r.slant -= wobble
	case motionFocused:
		l.gazeX += 3 * math.Sin(age*.7)
		r.gazeX = l.gazeX
	case motionExcited:
		bounce := .5 + .5*math.Sin(age*4.8)
		shake := 2.2*math.Sin(age*24) + .7*math.Sin(age*31)
		l.open += .04 * bounce
		r.open += .04 * bounce
		l.height += .05 * bounce
		r.height += .05 * bounce
		l.gazeY -= 1.5 * bounce
		r.gazeY -= 1.5 * bounce
		l.gazeX += shake
		r.gazeX += shake
		l.gazeY += .8 * math.Sin(age*29)
		r.gazeY += .8 * math.Sin(age*29)
	case motionNervous:
		dart := 7*math.Sin(age*5.5) + 2*math.Sin(age*2.4)
		l.gazeX += dart
		r.gazeX += dart
	case motionRelieved:
		l.open, r.open = relievedOpen(age), relievedOpen(age)
	default:
		l, r = animateExpressivePose(p, age)
	}
	dx, dy := poseGaze(p.motion, age)
	l.gazeX += dx
	r.gazeX += dx
	l.gazeY += dy
	r.gazeY += dy
	return l, r
}

func relievedOpen(age float64) float64 {
	phase := math.Mod(age, 5)
	switch {
	case phase < .35:
		return .2
	case phase < 1.2:
		return .2 + .5*(phase-.35)/.85
	case phase < 3.8:
		return .7
	default:
		return .7 - .5*(phase-3.8)/1.2
	}
}
