package fetch

import (
	"errors"
	"net"
	"net/http"
	"testing"
)

// FuzzBlockedIP feeds string IPs to the SSRF guard (GEN-24). Parsing
// never panics, and every loopback address stays blocked.
func FuzzBlockedIP(f *testing.F) {
	for _, seed := range []string{"127.0.0.1", "::1", "8.8.8.8", "169.254.169.254", "not-an-ip", "", "1.2.3.4.5"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		ip := net.ParseIP(s)
		if ip == nil {
			return
		}
		if ip.IsLoopback() && !blockedIP(ip) {
			t.Fatalf("loopback %s not blocked", s)
		}
	})
}

// FuzzMethodHeader feeds methods and headers to the guard (GEN-24).
// An empty method is rewritten to GET. Any other method off the
// allowlist, including one that contains a newline, is rejected.
func FuzzMethodHeader(f *testing.F) {
	for _, seed := range []string{"GET", "POST", "BREW", "", "get", "Cookie"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, method string) {
		err := checkMethodHeader(method, map[string]string{"X-Test": method})
		errKey := checkMethodHeader(method, map[string]string{method: "v"})
		if method == "" || methodAllowed(method) {
			if err != nil || errKey != nil {
				t.Fatalf("method %q: header err %v, key err %v", method, err, errKey)
			}
			return
		}
		if !errors.Is(err, ErrMethod) || !errors.Is(errKey, ErrMethod) {
			t.Fatalf("method %q: header err %v, key err %v", method, err, errKey)
		}
	})
}

// methodAllowed reports the methods checkMethodHeader accepts as written.
// Empty is not a member: the guard rewrites that to GET before the switch.
func methodAllowed(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
