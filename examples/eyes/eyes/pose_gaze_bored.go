package eyes

import "math"

func boredGaze(age float64) (float64, float64) {
	phase := math.Mod(age, boredPeriod)
	switch {
	case phase < boredPhaseA:
		return -boredRightX, 0
	case phase < boredPhaseB:
		k := smooth(phase - boredPhaseA)
		return -boredRightX + boredSweepX*k, 0
	case phase < boredPhaseC:
		return boredRightX, boredLevelY
	case phase < boredPhaseD:
		k := smooth(phase - boredPhaseC)
		return boredRightX - boredSweepX*k, boredLevelY
	default:
		return -boredRightX, 0
	}
}
