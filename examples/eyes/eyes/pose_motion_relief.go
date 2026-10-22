package eyes

import "math"

// relievedOpen eases the lids open, holds, and settles back over one period.
func relievedOpen(age float64) float64 {
	phase := math.Mod(age, relievedPeriod)
	switch {
	case phase < relievedPhaseA:
		return relievedClosed
	case phase < relievedPhaseB:
		return relievedClosed + relievedEaseAmp*(phase-relievedPhaseA)/relievedEaseLen
	case phase < relievedPhaseC:
		return relievedLidOpen
	default:
		return relievedLidOpen - relievedEaseAmp*(phase-relievedPhaseC)/relievedFallLen
	}
}
