package eyes

import "math"

// shyGaze walks a small loop: left, across, up, and back to rest.
func shyGaze(age float64) (float64, float64) {
	phase := math.Mod(age, shyPeriod)
	switch {
	case phase < shyPhaseA:
		return shyLeftX, shyLowY
	case phase < shyPhaseB:
		k := smooth((phase - shyPhaseA) / shyEase)
		return shyLeftX + shySweepX*k, shyLowY
	case phase < shyPhaseC:
		k := smooth((phase - shyPhaseB) / shyEase)
		return shyRightX - shySweepX*k, shyLowY + shySweepY*k
	case phase < shyPhaseD:
		return shyLeftX, shyRestY
	default:
		k := smooth(phase - shyPhaseD)
		return shyLeftX, shyRestY - shySweepY*k
	}
}
