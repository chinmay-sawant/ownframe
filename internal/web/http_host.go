package web

import (
	"net"
	"net/http"
	"strings"
)

// bindHostName returns the listen host as the configured name: a Host
// header for the address Serve bound is accepted alongside loopback.
func bindHostName(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return host
}

// allowedHost reports a Host header rooted at this machine.
// Loopback names pass; configuredName adds one app hostname.
func allowedHost(host, configuredName string) bool {
	name := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		name = h
	}
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.Trim(name, "[]")
	switch name {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return configuredName != "" && name == strings.ToLower(configuredName)
}

// requireLocalHost rejects DNS-rebinding Hosts before mux runs.
func requireLocalHost(configuredName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowedHost(r.Host, configuredName) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
