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
	if !sameOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)

		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)

	q := r.URL.Query()
	if len(q) == 0 {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if v := r.PostFormValue("x"); v != "" {
			q.Set("x", v)
		}
		if v := r.PostFormValue("y"); v != "" {
			q.Set("y", v)
		}
	}

	x, errX := strconv.ParseFloat(q.Get("x"), 64)
	y, errY := strconv.ParseFloat(q.Get("y"), 64)

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
