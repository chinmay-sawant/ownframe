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
	case motionConfused:
		l.gazeX += 8 * math.Sin(age*1.1)
		r.gazeX = l.gazeX
	case motionExcited:
		pulse := .025 * math.Sin(age*4)
		l.height += pulse
		r.height += pulse
	case motionShy:
		peek := 12 * math.Max(0, math.Sin(age*1.3))
		l.gazeX += peek
		r.gazeX += peek
	case motionNervous:
		dart := 13*math.Sin(age*7) + 7*math.Sin(age*3.1)
		l.gazeX += dart
		r.gazeX += dart
	case motionAnticipation:
		l.open, r.open = anticipationOpen(age), anticipationOpen(age)
	}
	return l, r
}

func anticipationOpen(age float64) float64 {
	phase := math.Mod(age, 4)
	if phase < .65 || phase >= 3.3 {
		return .05
	}
	if phase < 1.25 {
		return .05 + .9*(phase-.65)/.6
	}
	if phase < 2.8 {
		return .95
	}
	return .95 - .9*(phase-2.8)/.5
}
