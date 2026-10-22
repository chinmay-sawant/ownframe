//go:build !ownframe_debug

package web

import (
	"context"
	"net/http"
	"testing"
)

// TestDebugRoutesAbsentWithoutTag opts into every developer endpoint and
// checks the routes stay unregistered in a default build. The handlers stay
// compiled for direct unit tests; only the network registration behind
// -tags ownframe_debug is gone (GEN-15).
func TestDebugRoutesAbsentWithoutTag(t *testing.T) {
	t.Parallel()
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		opts := Options{InsecureAllowPprof: true, InsecureAllowDebugState: true}
		done <- ServeWithOptionsContext(ctx, debugApp(), addr, opts)
	}()
	waitUp(t, addr)
	if got := getStatus(t, addr, "/debug/state"); got != http.StatusNotFound {
		t.Fatalf("debug/state = %d, want 404 without ownframe_debug", got)
	}
	if got := getStatus(t, addr, "/debug/pprof/"); got != http.StatusNotFound {
		t.Fatalf("debug/pprof = %d, want 404 without ownframe_debug", got)
	}
}
