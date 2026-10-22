package store

import "time"

// Store-wide numbers, named for the mnd gate.
const (
	insertColsPerRow = 16
	exportDefaultCap = 8 << 20
	exportPageLimit  = 500
	ingestHoldDef    = 500 * time.Millisecond
	ingestPollDef    = 50 * time.Millisecond
	pageCap          = 500
	retainRows       = 200_000
	retainBytes      = 256 << 20
	sqlitePerm       = 3
	sqliteReadonly   = 8
	sqliteCantOpen   = 14
	storeDirPerm     = 0o700
	storeJobQueue    = 64
	storeDefTimeout  = 5 * time.Second
)
