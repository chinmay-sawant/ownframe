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
		glance := 12 * math.Sin(age*1.15)
		flutter := math.Max(0, math.Sin(age*4))
		l.gazeX += glance
		r.gazeX += glance
		l.gazeY += 1.5 * flutter
		r.gazeY += 1.5 * flutter
		l.open -= .16 * flutter
		r.open -= .16 * flutter
	case motionScared:
		shiver := math.Sin(age * 25)
		l.gazeX += 2.4 * shiver
		r.gazeX += 2.4 * shiver
		quiver := 1.8 * math.Sin(age*22+.8)
		l.gazeY += quiver
		r.gazeY += quiver
	case motionFrustrated:
		squeeze := .04 * math.Max(0, math.Sin(age*3.8))
		l.open -= squeeze
		r.open -= squeeze
	case motionSuspicious:
		glance := 5 * math.Sin(age*1.3)
		l.gazeX += glance
		r.gazeX += glance
	case motionDisgusted:
		recoil := math.Max(0, math.Sin(age*2.6))
		l.gazeX += 3 * recoil
		r.gazeX += 3 * recoil
		l.gazeY -= 1.5 * recoil
		r.gazeY -= 1.5 * recoil
		l.open -= .05 * recoil
		r.open -= .05 * recoil
	case motionProud:
		lift := 2.5 * math.Max(0, math.Sin(age*1.1))
		l.gazeY -= lift
		r.gazeY -= lift
		l.open += .02 * math.Max(0, math.Sin(age*.8))
		r.open += .02 * math.Max(0, math.Sin(age*.8))
	case motionSmug, motionMischievous, motionPleading:
		return animateReferenceExpression(p, age)
	}
	return l, r
}
