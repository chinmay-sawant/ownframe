package store

import "time"

// Store tuning constants.
const (
	dirPerm      = 0o750
	openTimeout  = 10 * time.Second
	retryDelay   = 5 * time.Millisecond
	maxLimit     = 1000
	scorePerLine = 1000
)
