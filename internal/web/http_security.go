package web

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
)

// securityHeaders sets nosniff and DENY framing on web responses.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// panicRecovery converts a panic into a 500 without ending Serve.
func panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("web handler panic: %s", sanitizeLine(fmt.Sprint(rec)))
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
