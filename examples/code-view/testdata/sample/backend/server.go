package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Server wires the API and the static dashboard.
type Server struct {
	cfg   Config
	store *Store
	log   *slog.Logger
	mux   *http.ServeMux
}

// NewServer returns the root handler.
func NewServer(cfg Config, store *Store, log *slog.Logger) *Server {
	s := &Server{cfg: cfg, store: store, log: log, mux: http.NewServeMux()}

	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/services", s.services)
	s.mux.HandleFunc("GET /api/checks", s.checks)
	s.mux.HandleFunc("GET /api/checks/{name}", s.history)
	s.mux.Handle("/", http.FileServer(http.Dir("../web")))

	return s
}

// ServeHTTP adds a request log line and the security headers.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("Cache-Control", "no-store")

	s.mux.ServeHTTP(w, r)

	if r.URL.Path != "/api/health" {
		s.log.Debug("request", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
	}
}

// health is the one route the load balancer sees.
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"services": len(s.cfg.Services),
		"interval": s.cfg.Interval.String(),
	})
}

// services returns the catalog.
func (s *Server) services(w http.ResponseWriter, _ *http.Request) {
	type row struct {
		Name string   `json:"name"`
		URL  string   `json:"url"`
		Tags []string `json:"tags"`
	}

	rows := make([]row, 0, len(s.cfg.Services))

	for _, svc := range s.cfg.Services {
		rows = append(rows, row{Name: svc.Name, URL: svc.URL, Tags: svc.Tags})
	}

	writeJSON(w, http.StatusOK, map[string]any{"services": rows})
}

// checks returns the latest result for every service.
func (s *Server) checks(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")

	latest, err := s.store.Latest(r.Context())
	if err != nil {
		s.fail(w, "latest", err)

		return
	}

	type row struct {
		Service  string `json:"service"`
		State    string `json:"state"`
		Status   int    `json:"status"`
		Latency  int64  `json:"latency_ms"`
		TTFB     int64  `json:"ttfb_ms"`
		At       string `json:"at"`
		Attempts int    `json:"attempts"`
	}

	rows := make([]row, 0, len(latest))

	for _, result := range latest {
		svc, ok := findService(s.cfg, result.Service)
		if !ok || (tag != "" && !hasTag(svc, tag)) {
			continue
		}

		rows = append(rows, row{
			Service:  result.Service,
			State:    result.State,
			Status:   result.Status,
			Latency:  result.Total.Milliseconds(),
			TTFB:     result.TTFB.Milliseconds(),
			At:       result.At.Format(time.RFC3339),
			Attempts: result.Attempts,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"checks":   rows,
		"updated":  time.Now().UTC().Format(time.RFC3339),
		"interval": s.cfg.Interval.String(),
	})
}

// history returns the last checks for one service.
func (s *Server) history(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	limit := 120
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 1440 {
			limit = n
		}
	}

	rows, err := s.store.History(r.Context(), name, limit)
	if err != nil {
		s.fail(w, "history", err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"service": name, "checks": rows})
}

// fail writes a JSON error and logs it.
func (s *Server) fail(w http.ResponseWriter, op string, err error) {
	s.log.Error("api", "op", op, "err", err)
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
}

// writeJSON encodes one response.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// findService looks up a catalog entry by name.
func findService(cfg Config, name string) (Service, bool) {
	for _, svc := range cfg.Services {
		if strings.EqualFold(svc.Name, name) {
			return svc, true
		}
	}

	return Service{}, false
}

// hasTag reports whether a service carries a tag.
func hasTag(svc Service, tag string) bool {
	for _, have := range svc.Tags {
		if strings.EqualFold(have, tag) {
			return true
		}
	}

	return false
}
