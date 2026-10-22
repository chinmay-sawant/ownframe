package transfer

import "time"

// Dummy transport, timeout, and permission tuning.
const (
	transferDirPerm   = 0o750
	transferFilePerm  = 0o600
	splitParts        = 2
	maxExtLen         = 16
	fakePace          = 25 * time.Millisecond
	fakeChunk         = 32 << 10
	fakeBaseSize      = 128 << 10
	fakeSizeSpan      = 2 << 20
	fakeUnknownMod    = 5
	fakeFailMod       = 7
	bodyHashBase      = 31
	bodyStep          = 7
	failNumer         = 3
	failDenom         = 5
	finalizeRetries   = 5
	finalizeBackoff   = 30 * time.Millisecond
	stallTimeout      = 30 * time.Second
	connectTimeout    = 10 * time.Second
	responseTimeout   = 20 * time.Second
	keepAliveInterval = 30 * time.Second
	maxIdleConns      = 4
	idleConnTimeout   = 90 * time.Second
	maxRedirects      = 10
	streamChunk       = 64 << 10
)
