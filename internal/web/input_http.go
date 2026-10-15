package web

import (
	"net/http"
)

func (s *server) typeText(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)

		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.app.Type(r.Context(), r.PostFormValue("text")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) backspace(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)

		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.app.Backspace(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
