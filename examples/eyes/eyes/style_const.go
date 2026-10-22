package eyes

// Base eye styles per pose: lid openness, lid share, slant, and gaze rest.
const (
	squintOpen      = 0.34
	squintLid       = 0.5
	angryOpen       = 0.64
	angryLid        = 0.08
	angrySlant      = 11
	angryGazeY      = -4
	relaxedOpen     = 0.68
	relaxedTop      = 0.82
	relaxedBottom   = 0.3
	relaxedGazeY    = 6
	sleepyOpen      = 0.42
	sleepyTop       = 0.9
	sleepyBottom    = 0.1
	sleepyGazeY     = 7
	surprisedHeight = 1.16
	surprisedGazeY  = -4
	curiousOpen     = 0.65
	curiousTop      = 0.6
	curiousSlantL   = -9
	curiousSlantR   = 4
	curiousGazeX    = 12
	curiousGazeY    = -5
	winkGazeX       = 8
	boredOpen       = 0.4
	boredTop        = 0.7
	boredBottom     = 0.3
	boredGazeY      = 10
)

// Sad, scared, and frustrated eye styles.
const (
	sadOpen         = 0.58
	sadTop          = 0.78
	sadBottom       = 0.2
	sadSlantL       = -5
	sadGazeY        = 10
	scaredHeight    = 1.16
	scaredTop       = 0.02
	scaredBottom    = 0.03
	scaredSlant     = 3
	scaredGazeY     = -2
	frustratedOpen  = 0.38
	frustratedGazeX = 5
	frustratedGazeY = 10
)
