package dump

// Scenario and percentile sizes for the headless dump.
const (
	scenarioWarmupTicks = 30
	scenarioFlapTicks   = 120
	scenarioIdleTicks   = 240
	scenarioResyncTicks = 150
	percentP50          = 50
	percentP95          = 95
	percentMax          = 100
	percentDenom        = 100
	percentRankAdjust   = 99
	pixelRound          = 0.5
	nanosPerMilli       = 1e6
	dumpDirPerm         = 0o750
	resyncMidAt         = 3
	dumpFilePerm        = 0o600
)
