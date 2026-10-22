package remote

import "time"

// UI and timing constants for the remote.
const (
	scrollDivisor      = 4
	maxJobs            = 64
	maxKeyParts        = 2
	noteQueueSize      = 8
	defaultFontSize    = 16
	defaultSensitivity = 100
	sensitivityScale   = 100
	halfDivisor        = 2
	pressFlashDelay    = 140 * time.Millisecond
	sensitivityStep    = 5
	percentScale       = 100
	sensitivityMin     = 25
	wakeDelay          = 8 * time.Second
	wakeMaxAttempts    = 2
	wakeRetryDelay     = 10 * time.Second
	repeatInitDelay    = 350 * time.Millisecond
	repeatFirstTick    = 120 * time.Millisecond
	repeatSecondAfter  = 2 * time.Second
	repeatThirdAfter   = 4 * time.Second
	repeatSecondTick   = 80 * time.Millisecond
	repeatThirdTick    = 50 * time.Millisecond
	sensitivityMax     = 300
	sensitivityRange   = 275
)
