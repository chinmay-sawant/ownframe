package store

import "time"

// Directory, file, and seed tuning for the store.
const (
	storeDirPerm     = 0o750
	storeFilePerm    = 0o600
	storeOpsQueue    = 64
	busyTimeout      = 5 * time.Second
	seedBaseSize     = 200 << 10
	seedSizeSpan     = 8 << 20
	seedSizeStep     = 7919
	seedETagMod      = 97
	seedTouchMinutes = 37
	seedAgeStepMin   = 91
	seedFailDiv      = 10
	seedCancelDiv    = 20
	seedQueuedDiv    = 12
	seedPausedDiv    = 14
	seedFailFrac     = 3
	seedCancelFrac   = 5
	seedHalfFrac     = 2
	seedUnknownMod   = 3
)
