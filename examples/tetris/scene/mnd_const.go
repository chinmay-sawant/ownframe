package scene

import "time"

// Scene tuning constants.
const (
	previewCap      = 16
	snapshotPoll    = 2 * time.Millisecond
	shortIDLen      = 8
	channelMax      = 255
	lightOverAlpha  = 0.94
	lightGhostAlpha = 0.34
	darkOverAlpha   = 0.95
	darkGhostAlpha  = 0.32
	previewMid      = 2
	previewMax      = 3
	workerQueue     = 8
	workerResults   = 32
)
