package eyes

import "math"

func animateExpressivePose(p eyePose, age float64) (eyeStyle, eyeStyle) {
	l, r := p.left, p.right
	switch p.motion {
	case motionSad:
		drift := 1.2 * math.Sin(age*1.4)
		l.gazeY += drift
		r.gazeY += drift
	case motionEmbarrassed:
		glance := 9 * math.Sin(age*2.4)
		l.gazeX += glance
		r.gazeX += glance
	case motionScared:
		dart := 14*math.Sin(age*8) + 4*math.Sin(age*15)
		l.gazeX += dart
		r.gazeX += dart
		pulse := .025 * math.Sin(age*6)
		l.open += pulse
		r.open += pulse
	case motionFrustrated:
		squeeze := .04 * math.Max(0, math.Sin(age*3.8))
		l.open -= squeeze
		r.open -= squeeze
	case motionSuspicious:
		glance := 5 * math.Sin(age*1.3)
		l.gazeX += glance
		r.gazeX += glance
	case motionSmug:
		l.open += .025 * math.Sin(age*1.2)
		r.gazeX += 2 * math.Sin(age*.9)
	case motionMischievous:
		l.open -= .12 * math.Max(0, math.Sin(age*1.7))
	case motionDisgusted:
		recoil := math.Max(0, math.Sin(age*2.6))
		l.gazeX -= 5 * recoil
		r.gazeX -= 5 * recoil
	case motionProud:
		lift := .025 * math.Sin(age*1.1)
		l.height += lift
		r.height += lift
	case motionPleading:
		appeal := .5 + .5*math.Sin(age*2.1)
		l.open += .04 * appeal
		r.open += .04 * appeal
		l.gazeY -= appeal
		r.gazeY -= appeal
	}
	return l, r
}
