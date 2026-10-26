package web

import (
	"net"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// one mutating token per 50ms; burst 40. maxInflight caps handlers;
// maxThrottleIP bounds the table. The burst leaves room for typing and
// wheel events while still bounding a scripted flood.
const (
	limitEvery    = 50 * time.Millisecond
	limitBurst    = 40
	maxInflight   = 8
	maxThrottleIP = 1024
	throttleKeep  = 5 * time.Minute
)

type throttleSeen struct {
	limiter *rate.Limiter
	last    time.Time
}

// allowAddr reports one token for the source IP; state lives on the server.
func (s *server) allowAddr(addr string) bool {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr
	}
	now := time.Now()
	s.limitMu.Lock()
	defer s.limitMu.Unlock()
	seen, ok := s.throttleHits[ip]
	if !ok {
		s.throttleHits[ip] = &throttleSeen{limiter: rate.NewLimiter(rate.Every(limitEvery), limitBurst), last: now}
		return true
	}
	seen.last = now
	if len(s.throttleHits) > maxThrottleIP {
		for key, item := range s.throttleHits {
			if now.Sub(item.last) > throttleKeep {
				delete(s.throttleHits, key)
			}
		}
	}
	return seen.limiter.Allow()
}

func mutatingPath(path string) bool {
	switch path {
	case "/click", "/type", "/backspace", "/key", "/wheel":
		return true
	}

	return false
}

// limit caps mutating POSTs per IP with 429 and in-flight with 503.
func (s *server) limit(next http.Handler) http.Handler {
	s.limitMu.Lock()
	if s.throttleHits == nil {
		s.throttleHits = map[string]*throttleSeen{}
	}
	if s.inflight == nil {
		s.inflight = make(chan struct{}, maxInflight)
	}
	s.limitMu.Unlock()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !mutatingPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		if !s.allowAddr(r.RemoteAddr) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		select {
		case s.inflight <- struct{}{}:
		default:
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		defer func() { <-s.inflight }()
		next.ServeHTTP(w, r)
	})
}
