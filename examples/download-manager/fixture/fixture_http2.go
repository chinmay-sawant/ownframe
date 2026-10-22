package fixture

import (
	"fmt"
	"net/http"
	"strconv"
)

// serveRange answers Range and If-Range; a changed validator gives 200.
func (s *Server) serveRange(w http.ResponseWriter, r *http.Request) {
	body := s.fill(requestSeed(r))
	w.Header().Set("ETag", etagOK)
	w.Header().Set("Accept-Ranges", "bytes")

	ifRange := r.Header.Get("If-Range")
	if ifRange != "" && ifRange != etagOK {
		writeBody(w, body, etagOK)

		return
	}

	if r.Header.Get("Range") == "" {
		writeBody(w, body, etagOK)

		return
	}

	start, end, ok := parseRange(r.Header.Get("Range"), int64(len(body)))
	if !ok {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", len(body)))
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)

		return
	}

	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(body)))
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.WriteHeader(http.StatusPartialContent)
	_, _ = w.Write(body[start : end+1])
}

// serveChanging bumps its ETag and body on every hit.
func (s *Server) serveChanging(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.round++
	round := s.round
	s.mu.Unlock()

	body := s.fill(round)
	writeBody(w, body, fmt.Sprintf(`"changing-%d"`, round))
}

// serveRedirect sends a 302 to the stable route.
func (s *Server) serveInterrupt(w http.ResponseWriter, r *http.Request) {
	body := s.fill(requestSeed(r))
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	_, _ = w.Write(body[:len(body)/2])
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	panic(http.ErrAbortHandler)
}
