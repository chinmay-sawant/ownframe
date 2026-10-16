package eyes

import "math"

func animateExpressivePose(p eyePose, age float64) (eyeStyle, eyeStyle) {
	l, r := p.left, p.right
	switch p.motion {
	case motionSad:
		drift := 1.2 * math.Sin(age*1.4)
		l.gazeY += drift
		r.gazeY += drift
	case motionScared:
		dart := 25*math.Sin(age*4.2) + 8*math.Sin(age*11.5)
		l.gazeX += dart
		r.gazeX += dart
		wide := .035 * math.Max(0, math.Sin(age*5.5))
		l.open += wide
		r.open += wide
	case motionFrustrated:
		drift := 15 * math.Sin(age*2.3)
		l.gazeX += drift
		r.gazeX += drift
		squeeze := .07 * math.Max(0, math.Sin(age*3.8))
		l.open -= squeeze
		r.open -= squeeze
	}
	return l, r
}
