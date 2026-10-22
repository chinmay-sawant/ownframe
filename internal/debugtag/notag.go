//go:build !ownframe_debug

// Package debugtag reports whether the ownframe_debug build tag is set.
// Default builds leave it off, so network-reachable debug handlers, pprof
// routes, and debug environment output stay out of the binary. Pass
// -tags ownframe_debug for a development build that keeps them.
package debugtag

// Enabled is false here: this file builds without ownframe_debug.
const Enabled = false
