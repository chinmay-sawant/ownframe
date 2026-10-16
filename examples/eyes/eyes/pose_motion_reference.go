package eyes

import "math"

func animateReferenceExpression(p eyePose, age float64) (eyeStyle, eyeStyle) {
	l, r := p.left, p.right
	switch p.motion {
	case motionSmug:
		glance := 4 + 2*math.Sin(age*.9)
		l.gazeX += glance
		r.gazeX += glance
		l.browUp += 1.5 * math.Sin(age*1.1)
		l.open -= .035 * math.Max(0, math.Sin(age*1.3))
	case motionMischievous:
		glance := 3 * math.Sin(age*1.2)
		l.gazeX += glance
		r.gazeX += glance
		wink := math.Max(0, math.Sin(age*1.6))
		l.open -= .17 * wink
		l.browUp += 2 * wink
		r.browUp += 1.2 * math.Max(0, math.Sin(age*1.6+.5))
	case motionPleading:
		appeal := .5 + .5*math.Sin(age*1.4)
		l.open += .02 * appeal
		r.open += .02 * appeal
		l.height += .015 * appeal
		r.height += .015 * appeal
		l.gazeY -= 2 * appeal
		r.gazeY -= 2 * appeal
		l.browUp += 3 * appeal
		r.browUp += 3 * appeal
	}
	return l, r
}
