package web

import (
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/chinmay-sawant/blinkless/layout"
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
	for _, k := range []string{"ox", "oy", "w", "h"} {
		if len(q.Get(k)) > 64 {
			http.Error(w, "bad coordinates", http.StatusBadRequest)
			return
		}
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
		if v := r.PostFormValue("id"); v != "" {
			q.Set("id", v)
		}
		if v := r.PostFormValue("action"); v != "" {
			q.Set("action", v)
		}
		for _, k := range []string{"ox", "oy", "w", "h"} {
			if v := r.PostFormValue(k); v != "" {
				q.Set(k, v)
			}
		}
	}
	if id := q.Get("id"); id != "" {
		if len(id) > maxDOMIDLength {
			http.Error(w, "bad coordinates", http.StatusBadRequest)
			return
		}
		for _, b := range s.app.Boxes() {
			if b.ID == id {
				q.Set("x", strconv.FormatFloat(b.X+b.W/2, 'f', -1, 64))
				q.Set("y", strconv.FormatFloat(b.Y+b.H/2, 'f', -1, 64))
				if nx, ny, ok := interpPoint(b, q); ok {
					q.Set("x", strconv.FormatFloat(nx, 'f', -1, 64))
					q.Set("y", strconv.FormatFloat(ny, 'f', -1, 64))
				}
				break
			}
		}
	} else if action := q.Get("action"); action != "" {
		if len(action) > maxDOMIDLength {
			http.Error(w, "bad coordinates", http.StatusBadRequest)
			return
		}
		for _, b := range s.app.Boxes() {
			if b.Action == action {
				q.Set("x", strconv.FormatFloat(b.X+b.W/2, 'f', -1, 64))
				q.Set("y", strconv.FormatFloat(b.Y+b.H/2, 'f', -1, 64))
				break
			}
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
	y += float64(s.scrollY)
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
	s.syncScroll()
	s.publishIfChangedLocked()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// interpPoint maps a browser click offset inside the clicked element into the
// matching box's CSS pixels, so the server caret lands where the browser put
// its caret instead of the box center. Missing or invalid offsets report
// false and the caller keeps the center.
func interpPoint(b layout.Box, q url.Values) (float64, float64, bool) {
	ox, errOX := strconv.ParseFloat(q.Get("ox"), 64)
	oy, errOY := strconv.ParseFloat(q.Get("oy"), 64)
	w, errW := strconv.ParseFloat(q.Get("w"), 64)
	h, errH := strconv.ParseFloat(q.Get("h"), 64)
	if errOX != nil || errOY != nil || errW != nil || errH != nil {
		return 0, 0, false
	}
	if math.IsNaN(ox) || math.IsInf(ox, 0) || math.IsNaN(oy) || math.IsInf(oy, 0) ||
		math.IsNaN(w) || math.IsInf(w, 0) || math.IsNaN(h) || math.IsInf(h, 0) {
		return 0, 0, false
	}
	if w <= 0 || h <= 0 || ox < 0 || oy < 0 || ox > w || oy > h {
		return 0, 0, false
	}

	return b.X + ox/w*b.W, b.Y + oy/h*b.H, true
}
