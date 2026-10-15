package web

import (
	"log"
	"net/http"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// pdfScreen names the PDF route's screen. Deprecated: use
// host.PDFExporter[page.PDFOptions]; kept so ServeWithOptions reads unchanged.
type pdfScreen = host.PDFExporter[page.PDFOptions]

func (s *server) pdf(w http.ResponseWriter, r *http.Request) {
	exporter, ok := s.app.(host.PDFExporter[page.PDFOptions])
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := exporter.PDF(r.Context(), page.PDFOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Cache-Control", "no-store")

	if _, err := w.Write(data); err != nil {
		log.Printf("write pdf: %v", err)
	}
}
