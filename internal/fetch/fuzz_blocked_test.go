package fetch

import (
	"net"
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
func FuzzMethodHeader(f *testing.F) {
	for _, seed := range []string{"GET", "POST", "BREW", "", "get", "Cookie"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, method string) {
		_ = checkMethodHeader(method, map[string]string{"X-Test": method})
		_ = checkMethodHeader(method, map[string]string{method: "v"})
	})
}
