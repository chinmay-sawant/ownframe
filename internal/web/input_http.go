package web

import (
	"context"
	"math"
	"net/http"
	"strconv"
)

// selector is the page method that places the caret from a click point.
type selector interface {
	SelectAt(ctx context.Context, x, y float64) error
}

// click places the caret before it runs the click handler, so a handler that
// focuses or selects a field is not blurred by the caret placement.
func (s *server) click(w http.ResponseWriter, r *http.Request) {
	x, errX := strconv.ParseFloat(r.URL.Query().Get("x"), 64)
	y, errY := strconv.ParseFloat(r.URL.Query().Get("y"), 64)

	if errX != nil || errY != nil || math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
		http.Error(w, "bad coordinates", http.StatusBadRequest)

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if sel, ok := s.app.(selector); ok {
		if err := sel.SelectAt(r.Context(), x, y); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	}

	if err := s.app.Click(r.Context(), x, y); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) typeText(w http.ResponseWriter, r *http.Request) {
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.app.Backspace(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
