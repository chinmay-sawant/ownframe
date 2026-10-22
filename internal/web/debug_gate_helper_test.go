package web

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// debugApp is the minimal app the route-gate tests serve: a display list
// the debug handler can serialize without touching a real page.
func debugApp() *fakeDebug {
	return &fakeDebug{display: &layout.Display{Width: 10, Height: 10}}
}

// debugTestClient never uses the shared global client: timeouts stay
// explicit and global client state cannot change debug-gate fetches.
func debugTestClient() *http.Client {
	return &http.Client{Timeout: 3 * time.Second}
}

// getStatus starts a server with every developer endpoint opted in and
// returns the status of one GET.
func getStatus(t *testing.T, addr, path string) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := debugTestClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode
}
