package eyes

import "math"

func mixStyle(from, to eyeStyle, k float64) eyeStyle {
	return eyeStyle{
		open:   from.open + (to.open-from.open)*k,
		height: from.height + (to.height-from.height)*k,
		top:    from.top + (to.top-from.top)*k,
		bottom: from.bottom + (to.bottom-from.bottom)*k,
		slant:  from.slant + (to.slant-from.slant)*k,
		gazeX:  from.gazeX + (to.gazeX-from.gazeX)*k,
		gazeY:  from.gazeY + (to.gazeY-from.gazeY)*k,
	}
}

func (a *App) stepPose(dt float64) {
	s := &a.st
	pose := poseFor(s.mood)
	k := 1 - math.Exp(-dt*4.5)
	s.styleL = mixStyle(s.styleL, pose.left, k)
	s.styleR = mixStyle(s.styleR, pose.right, k)
	s.moodL, s.moodR = s.styleL.open, s.styleR.open
}
