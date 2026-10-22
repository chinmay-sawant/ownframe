package eyes

// Shy gaze walks a small loop: left, across, up, and back to rest.
const (
	shyPeriod = 4.8
	shyPhaseA = 1.2
	shyPhaseB = 2
	shyPhaseC = 2.8
	shyPhaseD = 3.8
	shyEase   = 0.8
	shyLeftX  = -7
	shyRightX = 5
	shyLowY   = -7
	shyRestY  = 6
	shySweepX = 12
	shySweepY = 13
)

// Bored gaze drifts sideways, holds, and drifts back.
const (
	boredPeriod = 6
	boredPhaseA = 1.5
	boredPhaseB = 2.5
	boredPhaseC = 3.7
	boredPhaseD = 4.7
	boredRightX = 8
	boredLevelY = -1
	boredSweepX = 16
)

// smoothstep coefficients for eased gaze segments.
const (
	smoothHigh = 3
	smoothLow  = 2
)
