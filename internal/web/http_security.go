package web

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// securityHeaders sets nosniff and a simple anti-framing header on every
// response from the web preview.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// panicRecovery converts a panic in any handler into a 500 response instead
// of crashing the Serve process.
func panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("web handler panic: %v", rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// validateBindAddr rejects addresses that would expose Serve on every
// interface or that lack a host or port.
func validateBindAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	if host == "" || port == "" {
		return errors.New("ownframe: bind address must include a non-empty host and port")
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		return errors.New("ownframe: bind address must not use an unspecified or wildcard IP")
	}

	if host == "0.0.0.0" || host == "::" || host == "[::]" {
		return errors.New("ownframe: bind address must not use an unspecified or wildcard IP")
	}

	return nil
}

// sameOrigin reports whether request r's Origin header, when present, refers to
// the same host as r. A missing Origin header is permitted.
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}

	return strings.EqualFold(u.Host, r.Host)
}
