package cat

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

// withHostCheck rejects requests with non-loopback Host headers to protect
// against DNS rebinding attacks.
func withHostCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkHost(r) {
			reply(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// checkHost validates that r.Host strictly targets a loopback interface
// (localhost, 127.0.0.1, [::1], or loopback IP).
func checkHost(r *http.Request) bool {
	return isLoopbackHost(r.Host)
}

func isLoopbackHost(hostHeader string) bool {
	if hostHeader == "" {
		return false
	}
	host := hostHeader
	if h, _, err := net.SplitHostPort(hostHeader); err == nil {
		host = h
	}
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// checkOrigin fails closed: a missing Origin is rejected on gated
// routes. Curl callers send an explicit Origin header matching the
// request scheme and host.
func checkOrigin(r *http.Request) bool {
	if !checkHost(r) {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if !strings.EqualFold(u.Scheme, scheme) {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}
