package web

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"net/http"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
)

func (s *server) page(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reload(r.Context())
	viewW, viewH := s.app.Size()
	data := shellData{
		Areas:      areas(s.app),
		Reload:     s.watching(),
		Width:      viewW,
		Height:     viewH,
		Generation: s.app.Generation() + s.frameGen,
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
	var pngBytes []byte
	if s.app.Display() != nil {
		viewW, h := s.app.Size()
		gen := s.app.Generation() + s.frameGen
		if s.viewportPNG != nil && s.viewportGen == gen && s.viewportWidth == viewW && s.viewportHeight == h {
			pngBytes = s.viewportPNG
		} else {
			img, err := bitmap.Viewport(s.app.Display(), image.Rect(0, s.scrollY, viewW, s.scrollY+h), 1)
			if err != nil {
				http.Error(w, "render failed", http.StatusInternalServerError)
				return
			}
			var buf bytes.Buffer
			if err = png.Encode(&buf, img); err != nil {
				http.Error(w, "encode failed", http.StatusInternalServerError)
				return
			}
			pngBytes = buf.Bytes()
			s.viewportPNG = pngBytes
			s.viewportGen = gen
			s.viewportWidth = viewW
			s.viewportHeight = h
		}
	} else {
		pngBytes = s.app.PNG()
	}
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
