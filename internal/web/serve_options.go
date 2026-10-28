package web

import (
	"net"
	"time"
)

// Options tunes ServeWithOptions. The zero value is safe: no developer
// endpoints are registered and only loopback binds are accepted. Risky
// opt-ins carry an InsecureAllow prefix so call sites read risky.
type Options struct {
	// Listener supplies a pre-bound network listener for Serve. When set,
	// Serve uses Listener instead of opening a new TCP listener on addr.
	// This avoids port races in tests and embeds into custom server setups.
	Listener net.Listener
	// Deprecated: use InsecureAllowPprof. Either field registers the
	// GET /debug/pprof/* endpoints.
	Perf bool
	// Deprecated: use InsecureAllowDebugState. Either field registers
	// GET /debug/state (and implies inspector access).
	Debug bool
	// InsecureAllowPprof registers GET /debug/pprof/* for CPU, heap,
	// goroutine, and trace profiles. The name marks the exposure.
	InsecureAllowPprof bool
	// InsecureAllowDebugState registers GET /debug/state, which exposes
	// layout boxes with field text, stats, and runtime snapshot. Keep
	// it off before any non-loopback bind.
	InsecureAllowDebugState bool
	// InsecureAllowRemoteBind permits a non-loopback listen address.
	// Without it Serve accepts loopback binds only: plaintext HTTP plus
	// a LAN listener would expose clicks, typing, and debug state.
	InsecureAllowRemoteBind bool
}

// maxFormBytes caps the POST form bodies the preview handlers read.
const maxFormBytes = 256 * 1024

// maxHeaderBytes caps request headers at 64 KiB (WEB-4).
const maxHeaderBytes = 1 << 16

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)
