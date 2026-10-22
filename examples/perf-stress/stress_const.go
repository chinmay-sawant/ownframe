package main

// Stress dashboard geometry and tuning, in CSS pixels unless noted.
const (
	stressWidth   = 1280
	stressHeight  = 900
	stressMinW    = 640
	stressMinH    = 480
	stressRows    = 280
	progressStep  = 10
	progressCycle = 101
	progressScale = 100
	// eqMid is both the equalizer midpoint and its amplitude.
	eqMid        = 0.5
	eqFloor      = 0.2
	eqSpan       = 0.8
	eqPeriodBase = 120
	eqPeriodStep = 35
	eqPhaseStep  = 1.9
	orderBase    = 1000
	deltaFactor  = 7
	deltaMod     = 99
)
