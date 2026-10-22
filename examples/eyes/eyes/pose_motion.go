package eyes

import "math"

func animatePose(p eyePose, age float64) (eyeStyle, eyeStyle) {
	l, r := p.left, p.right
	wave := math.Sin(age * relaxedWave)
	switch p.motion {
	case motionRelaxed:
		l.open += relaxedOpenEase * wave
		r.open += relaxedOpenEase * wave
	case motionHappy:
		bright := math.Max(0, math.Sin(age*happyBrightFreq))
		grin := math.Max(0, math.Sin(age*happyBrightFreq-happyGrinLag))
		l.open += happyBrightOpen*bright - happyGrinClose*grin
		l.top -= happyBrightTop * bright
		l.bottom += happyGrinBottom * grin
		l.gazeY -= happyGazeBright * bright
		r.bottom = l.bottom
		r.open, r.top, r.gazeY = l.open, l.top, l.gazeY
	case motionWorried:
		wobble := worriedWobbleAmp * math.Sin(age*worriedWobbleFreq)
		l.slant += wobble
		r.slant -= wobble
	case motionFocused:
		l.gazeX += focusedGlanceAmp * math.Sin(age*focusedGlanceFreq)
		r.gazeX = l.gazeX
	case motionExcited:
		bounce := excitedBounce + excitedBounce*math.Sin(age*excitedBounceFreq)
		shake := excitedShakeAmp*math.Sin(age*excitedShakeFreq) + excitedShakeAmp2*math.Sin(age*excitedShakeFreq2)
		l.open += excitedOpenEase * bounce
		r.open += excitedOpenEase * bounce
		l.height += excitedHeightEase * bounce
		r.height += excitedHeightEase * bounce
		l.gazeY -= excitedGazeDip * bounce
		r.gazeY -= excitedGazeDip * bounce
		l.gazeX += shake
		r.gazeX += shake
		l.gazeY += excitedTremorAmp * math.Sin(age*excitedTremorFreq)
		r.gazeY += excitedTremorAmp * math.Sin(age*excitedTremorFreq)
	case motionNervous:
		dart := nervousDartAmp*math.Sin(age*nervousDartFreq) + nervousDartAmp2*math.Sin(age*nervousDartFreq2)
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
