package depthalert

// Window and display tuning for the depth-alert page.
const (
	windowWidth  = 360
	windowHeight = 280
	windowMinW   = 320
	windowMinH   = 260
)

// missingFrames clears the view after this many untracked frames.
const missingFrames = 3

// settleFrames accepts a new state after this many matching frames.
const settleFrames = 3

// State severities, worst first.
const (
	severityStopNow = 3
	severityStop    = 2
	severityUnknown = 1
)

// Deadbands keep tiny sensor jitter off the view.
const (
	nearDeadband  = 200 // millimetres
	coverDeadband = 8   // coverage percent
)

// coverBucket rounds coverage down to a multiple for display.
const coverBucket = 5

// mmPerM renders millimetres as metres with one decimal.
const mmPerM = 1000
