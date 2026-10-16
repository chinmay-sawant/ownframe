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
		l.bottom -= .06 * math.Sin(age*math.Pi)
		r.bottom = l.bottom
	case motionWorried:
		wobble := 1.4 * math.Sin(age*2.1)
		l.slant += wobble
		r.slant -= wobble
	case motionFocused:
		l.gazeX += 3 * math.Sin(age*.7)
		r.gazeX = l.gazeX
	case motionExcited:
		bounce := .5 + .5*math.Sin(age*4.8)
		l.open += .04 * bounce
		r.open += .04 * bounce
		l.height += .05 * bounce
		r.height += .05 * bounce
		l.gazeY -= 1.5 * bounce
		r.gazeY -= 1.5 * bounce
	case motionShy:
		peek := 12 * math.Max(0, math.Sin(age*1.3))
		l.gazeX += peek
		r.gazeX += peek
	case motionNervous:
		dart := 13*math.Sin(age*7) + 7*math.Sin(age*3.1)
		l.gazeX += dart
		r.gazeX += dart
	case motionRelieved:
		l.open, r.open = relievedOpen(age), relievedOpen(age)
	default:
		l, r = animateExpressivePose(p, age)
	}
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
