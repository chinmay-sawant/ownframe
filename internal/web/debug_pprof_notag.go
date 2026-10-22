//go:build !ownframe_debug

package web

import "net/http"

// registerPprof is a no-op without ownframe_debug. The pprof import stays
// out of release binaries, and /debug/pprof/* answers 404 there.
func registerPprof(_ *http.ServeMux) {}
