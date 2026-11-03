//go:build ownframe_debug

package window

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"
)

const (
	debugReadHeaderTimeout = 5 * time.Second
	debugReadTimeout       = 15 * time.Second
	debugWriteTimeout      = 15 * time.Second
	debugIdleTimeout       = 60 * time.Second
	debugMaxHeaderBytes    = 1 << 16
)

// startDebugServer listens on a loopback address and serves the window's
// profiling snapshot with the Go profiler. It runs beside the event loop:
// the game thread publishes immutable copies, the handler only loads them.
// A nil game or an empty address is a programmer error and returns an error.
// A non-loopback address is rejected without a remote opt-in: forward over
// ssh when off the machine. It returns the bound address for logs.
func startDebugServer(ctx context.Context, game *shell, addr string) (string, func(), error) {
	if ctx == nil {
		return "", nil, errors.New("window: nil context")
	}
	if game == nil {
		return "", nil, errors.New("window: nil game")
	}
	if addr == "" {
		return "", nil, errors.New("window: debug address must include a non-empty host and port")
	}
	if err := checkDebugBindAddr(addr); err != nil {
		return "", nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /debug/state", func(w http.ResponseWriter, _ *http.Request) {
		serveWindowDebugState(game, w)
	})
	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return "", nil, err
	}
	bound := ln.Addr().String()
	game.debugEnabled = true
	handler := debugPanicRecovery(debugSecurityHeaders(debugRequireLocalHost(debugBindHostName(bound), mux)))
	hs := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: debugReadHeaderTimeout,
		ReadTimeout:       debugReadTimeout,
		WriteTimeout:      debugWriteTimeout,
		IdleTimeout:       debugIdleTimeout,
		MaxHeaderBytes:    debugMaxHeaderBytes,
	}
	go func() {
		if err := hs.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("window debug serve: %s", sanitizeDebugLine(err.Error()))
		}
	}()
	stop := context.AfterFunc(ctx, func() { _ = hs.Shutdown(context.Background()) })
	log.Printf("window: debug listening on http://%s/debug/state", bound)
	return bound, func() {
		stop()
		_ = hs.Shutdown(context.Background())
	}, nil
}

// serveWindowDebugState answers one snapshot the game thread published. It
// touches no page: a 503 before the first publish beats a data race.
func serveWindowDebugState(game *shell, w http.ResponseWriter) {
	snap := game.debugSnapshot.Load()
	if snap == nil {
		http.Error(w, "no snapshot yet", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(snap); err != nil {
		log.Printf("window: write debug state: %s", sanitizeDebugLine(err.Error()))
	}
}

// checkDebugBindAddr keeps the wildcard rejection and the loopback rule:
// only localhost and loopback IPs bind. There is no remote opt-in.
func checkDebugBindAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if host == "" || port == "" {
		return errors.New("window: debug address must include a non-empty host and port")
	}
	trimmed := strings.Trim(host, "[]")
	if ip := net.ParseIP(trimmed); ip != nil && ip.IsUnspecified() {
		return errors.New("window: debug address must not use an unspecified or wildcard IP")
	}
	if host == "0.0.0.0" || host == "::" || host == "[::]" {
		return errors.New("window: debug address must not use an unspecified or wildcard IP")
	}
	if strings.ToLower(host) == "localhost" {
		return nil
	}
	if ip := net.ParseIP(trimmed); ip == nil || !ip.IsLoopback() {
		return errors.New("window: debug address must be loopback (127.0.0.1, ::1, or localhost)")
	}
	return nil
}

// debugBindHostName returns the listen host for the Host allowlist.
func debugBindHostName(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return host
}

// debugAllowedHost passes loopback request hosts and the configured name.
func debugAllowedHost(host, configuredName string) bool {
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

// debugRequireLocalHost rejects DNS-rebinding Hosts before the mux runs.
func debugRequireLocalHost(configuredName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !debugAllowedHost(r.Host, configuredName) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// debugSecurityHeaders sets the same framing and sniffing guards as web.
func debugSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

// debugPanicRecovery converts a handler panic into a 500 without ending Serve.
func debugPanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("window debug handler panic: %s", sanitizeDebugLine(fmt.Sprint(rec)))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
