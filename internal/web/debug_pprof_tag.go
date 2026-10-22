//go:build ownframe_debug

package web

import (
	"net/http"
	"net/http/pprof"
)

// registerPprof mounts the pprof handlers. This file only builds with
// -tags ownframe_debug, so release binaries never link net/http/pprof.
func registerPprof(mux *http.ServeMux) {
	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
}
