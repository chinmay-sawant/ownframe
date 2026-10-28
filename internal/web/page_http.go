package web

import (
	"bytes"
	"log"
	"net/http"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (s *server) page(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reload(r.Context())
	front := host.SelectFront(s.app)
	data := shellData{
		Areas:      areas(s.app),
		Reload:     s.watching(),
		Width:      front.Width,
		Height:     front.Height,
		Generation: front.Generation,
	}
	var buf bytes.Buffer
	if err := s.shell.Execute(&buf, data); err != nil {
		log.Printf("page: %s", sanitizeLine(err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("write page: %s", sanitizeLine(err.Error()))
	}
}

func (s *server) frame(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reload(r.Context())
	pngBytes := s.app.PNG()
	if len(pngBytes) == 0 {
		http.Error(w, "no frame", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(pngBytes); err != nil {
		log.Printf("write png: %s", sanitizeLine(err.Error()))
	}
}
