package fixture

import "time"

// Deterministic body and hash sizes for the test fixture.
const (
	defaultBodySize  = 64 << 10
	slowChunk        = 1024
	chunkedChunk     = 4096
	hashBase         = 31
	hashMask         = 0xffff
	decimalBase      = 10
	bodyStep         = 7
	defaultSlowPause = 25 * time.Millisecond
)
