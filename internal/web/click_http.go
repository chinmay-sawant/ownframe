package web

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// click places the caret before the click handler runs, so a handler
// that focuses a field is not blurred by caret placement.
func (s *server) click(w http.ResponseWriter, r *http.Request) {
	if !requireOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
	q := r.URL.Query()
	n := 0
	for _, v := range q {
		n += len(v)
	}
	if len(q) > 16 || n > 16 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(q.Get("x")) > 64 || len(q.Get("y")) > 64 {
		http.Error(w, "bad coordinates", http.StatusBadRequest)
		return
	}
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
	err := host.Batch(r.Context(), s.app, func() error {
		if sel, ok := s.app.(host.Selector); ok {
			if err := sel.SelectAt(r.Context(), x, y); err != nil {
				return err
			}
		}
		return s.ptr().Click(r.Context(), x, y)
	})
	if err != nil {
		log.Printf("click: %s", sanitizeLine(err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
