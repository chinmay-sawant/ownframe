// Package fixture serves deterministic test downloads over localhost with
// net/http/httptest: complete, slow, chunked, range-capable, changing,
// redirecting, and interrupting responses. Tests and the example's real
// mode can point at it so no external network is needed.
package fixture

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Routes the fixture serves.
const (
	PathOK        = "/ok"
	PathSlow      = "/slow"
	PathChunked   = "/chunked"
	PathRange     = "/range"
	PathChanging  = "/changing"
	PathRedirect  = "/redirect"
	PathInterrupt = "/interrupt"
)

// etagOK is the validator for the stable routes.
const etagOK = `"fixture-v1"`

// Server is a running fixture service.
type Server struct {
	*httptest.Server
	// Size is the generated body length. Zero means 64 KiB.
	Size int64
	// SlowPause is the delay between slow chunks. Zero means 25 ms.
	SlowPause time.Duration

	mu    sync.Mutex
	round int // hit count for /changing
}

// NewServer starts the fixture on a loopback port.
func NewServer() *Server {
	s := &Server{Size: defaultBodySize, SlowPause: defaultSlowPause}

	mux := http.NewServeMux()
	mux.HandleFunc(PathOK, s.serveOK)
	mux.HandleFunc(PathSlow, s.serveSlow)
	mux.HandleFunc(PathChunked, s.serveChunked)
	mux.HandleFunc(PathRange, s.serveRange)
	mux.HandleFunc(PathChanging, s.serveChanging)
	mux.HandleFunc(PathRedirect, s.serveRedirect)
	mux.HandleFunc(PathInterrupt, s.serveInterrupt)
	mux.HandleFunc("/status", s.serveStatus)

	s.Server = httptest.NewServer(mux)

	return s
}

// Close stops the server.
func (s *Server) Close() { s.Server.Close() }

// fill writes the deterministic body for a version seed.
func (s *Server) fill(seed int) []byte {
	size := s.Size
	if size <= 0 {
		size = defaultBodySize
	}

	out := make([]byte, size)
	for i := range out {
		out[i] = byte(seed*hashBase + i*bodyStep) //nolint:gosec // reason: seed and index bounded, test body fits a byte.
	}

	return out
}
