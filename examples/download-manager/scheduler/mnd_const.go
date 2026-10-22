package scheduler

import "time"

// Default engine cadence.
const (
	defaultProgressEvery  = 500 * time.Millisecond
	defaultShutdownBudget = 5 * time.Second
)
