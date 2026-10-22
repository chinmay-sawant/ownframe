package eyes

import "math"

func animateExpressivePose(p eyePose, age float64) (eyeStyle, eyeStyle) {
	l, r := p.left, p.right
	switch p.motion {
	case motionSad:
		drift := sadDriftAmp * math.Sin(age*sadDriftFreq)
		l.gazeY += drift
		r.gazeY += drift
	case motionScared:
		dart := scaredDartAmp*math.Sin(age*scaredDartFreq) + scaredDartAmp2*math.Sin(age*scaredDartFreq2)
		l.gazeX += dart
		r.gazeX += dart
		wide := scaredWideAmp * math.Max(0, math.Sin(age*scaredWideFreq))
		l.open += wide
		r.open += wide
	case motionFrustrated:
		drift := frustratedDriftAmp * math.Sin(age*frustratedDriftFreq)
		l.gazeX += drift
		r.gazeX += drift
		squeeze := frustratedSqueeze * math.Max(0, math.Sin(age*frustratedSqueezeAt))
		l.open -= squeeze
		r.open -= squeeze
	}
	return l, r
}
