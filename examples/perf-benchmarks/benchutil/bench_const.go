package benchutil

// Frame tuning: fallback step, clamp, and half gap for the tick painter.
const (
	flapFrameRate = 60.0
	flapMaxDt     = 0.1
	flapHalfGap   = flapGap / 2
)

// Reset pipes: the two visible pipes a reset starts with, in CSS pixels.
const (
	resetPipeX1   = 210.0
	resetPipeGap1 = 300.0
	resetPipeX2   = 390.0
	resetPipeGap2 = 220.0
)

// Window geometry for the three benchmark demos, in CSS pixels.
const (
	flapWidth       = 480
	flapHeight      = 720
	flapMinWidth    = 320
	flapMinHeight   = 480
	largeWidth      = 800
	largeHeight     = 600
	largeMinWidth   = 320
	largeMinHeight  = 240
	normalWidth     = 480
	normalHeight    = 640
	normalMinWidth  = 320
	normalMinHeight = 480
)
