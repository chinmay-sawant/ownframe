//go:build ownframe_debug

package web

import (
	"context"
	"net/http"
	"testing"
)

// TestDebugRoutesPresentWithTag starts the same opted-in server as
// TestDebugRoutesAbsentWithoutTag and checks the routes answer with
// -tags ownframe_debug, so the gate test above stays honest (GEN-15).
func TestDebugRoutesPresentWithTag(t *testing.T) {
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
	if got := getStatus(t, addr, "/debug/state"); got != http.StatusOK {
		t.Fatalf("debug/state = %d, want 200 with ownframe_debug", got)
	}
	if got := getStatus(t, addr, "/debug/pprof/"); got != http.StatusOK {
		t.Fatalf("debug/pprof = %d, want 200 with ownframe_debug", got)
	}
}
