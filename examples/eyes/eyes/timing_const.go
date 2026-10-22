package eyes

import (
	"math"
	"time"
)

// Mood pacing: the spiral holds briefly, other moods linger.
const (
	spiralMoodDur = 4 * time.Second
	autoMoodDur   = 8 * time.Second
)

// Tilt and shake settle back on the frame loop.
const (
	tiltHold      = 1500 * time.Millisecond
	shakeMoodHold = 650 * time.Millisecond
)

// Spiral easing and spin in milliseconds, rate, and degrees.
const (
	spiralSettle = 500 * time.Millisecond
	spiralRate   = 12
	spiralMinMix = 0.001
	spiralTurn   = 360
	spiralSpin   = 600
)

// Saccade scheduling and blink timing in milliseconds and seconds.
const (
	saccadeDelayMin = 1400
	saccadeDelayVar = 1600
	blinkDelayMin   = 2200
	blinkDelayVar   = 2600
	blinkPeriod     = 0.22
	blinkInitial    = 2 * time.Second
	baseFPS         = 60
	maxFrameDT      = 0.1
)

// Lid and breathing response per frame.
const (
	lidClose  = 0.98
	lookRate  = 18
	poseRate  = 4.5
	breathAmp = 0.014
	breathHz  = 0.35
	twoPi     = 2 * math.Pi
)

// Lid geometry in CSS pixels and the hidden parking spot.
const (
	lidTopShare  = 0.6
	lidLowShare  = 0.4
	lidMinGap    = 6
	lidOutset    = 2
	lidGrow      = 4
	minLidH      = 3
	hiddenCoord  = -1000
	eyeAlphaMin  = 0.0001
	angrySwayAmp = 2
	angrySwayHz  = 2.4
	eyeRound     = 0.98
	spiralBase   = 0.82
	spiralPulse  = 0.24
	radPerDeg    = math.Pi / 180
)

// Wink cycle: full period, edge ramps, and the hold between them.
const (
	winkCycle = 2400 * time.Millisecond
	winkEdge  = 240 * time.Millisecond
	winkHold  = 420 * time.Millisecond
)
