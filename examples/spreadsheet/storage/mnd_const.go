package storage

import "time"

// Storage bounds, named for the mnd gate.
const (
	rangeDefLimit = 5000
	rangeMaxLimit = 50000
	storeDirPerm  = 0o700
	storeQueueCap = 64
	storeTimeout  = 10 * time.Second
)
