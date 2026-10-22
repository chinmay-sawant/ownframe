package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSec_WebSecurityHeadersSet pins the hardening headers.
func TestSec_WebSecurityHeadersSet(t *testing.T) {
	t.Parallel()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	rec := httptest.NewRecorder()
	securityHeaders(next).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), "GET", "/", nil))
	for k, want := range map[string]string{"X-Content-Type-Options": "nosniff", "X-Frame-Options": "DENY", "Referrer-Policy": "no-referrer"} {
		if rec.Header().Get(k) != want {
			t.Fatalf("%s = %q", k, rec.Header().Get(k))
		}
	}
	if rec.Header().Get("Permissions-Policy") == "" {
		t.Fatal("missing Permissions-Policy")
	}
}

// TestSec_WebDevSurfacesOffByDefault pins secure-by-default options:
// no pprof or state without opt-in, loopback binds only without the
// remote-bind opt-in.
func TestSec_WebDevSurfacesOffByDefault(t *testing.T) {
	t.Parallel()
	var zero Options
	if effectivePprof(zero) || effectiveDebug(zero) {
		t.Fatal("developer endpoints on by default")
	}
	if err := checkBindAddr("192.168.1.9:8080", zero); !errors.Is(err, errRemoteBind) {
		t.Fatalf("remote bind: %v", err)
	}
	for _, addr := range []string{"127.0.0.1:8080", "localhost:8080", "[::1]:8080"} {
		if err := checkBindAddr(addr, zero); err != nil {
			t.Fatalf("%s: %v", addr, err)
		}
	}
	if err := checkBindAddr("192.168.1.9:8080", Options{InsecureAllowRemoteBind: true}); err != nil {
		t.Fatalf("opt-in remote: %v", err)
	}
}
