package fixture

import (
	"net/http"
	"strconv"
	"time"
)

// serveOK sends a complete response with a length and a validator.
func (s *Server) serveOK(w http.ResponseWriter, r *http.Request) {
	body := s.fill(requestSeed(r))
	w.Header().Set("ETag", etagOK)
	w.Header().Set("Last-Modified", "Wed, 01 Oct 2025 12:00:00 GMT")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	_, _ = w.Write(body)
}

// serveSlow trickles the body in small chunks.
func (s *Server) serveSlow(w http.ResponseWriter, r *http.Request) {
	body := s.fill(requestSeed(r))
	w.Header().Set("ETag", `"fixture-slow"`)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))

	for off := 0; off < len(body); off += slowChunk {
		end := off + slowChunk
		if end > len(body) {
			end = len(body)
		}

		_, _ = w.Write(body[off:end])
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		if s.SlowPause > 0 {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(s.SlowPause):
			}
		}
	}
}

// serveChunked hides the length by using chunked transfer encoding.
func (s *Server) serveChunked(w http.ResponseWriter, r *http.Request) {
	body := s.fill(requestSeed(r))
	w.Header().Set("ETag", `"fixture-chunked"`)

	for off := 0; off < len(body); off += chunkedChunk {
		end := off + chunkedChunk
		if end > len(body) {
			end = len(body)
		}

		_, _ = w.Write(body[off:end])
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}
