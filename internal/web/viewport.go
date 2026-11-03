package web

import (
	"math"
	"net/http"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// scrollStep is the CSS pixels scrolled per wheel delta unit.
const scrollStep = 48

func (s *server) scrollPage(dy float64) { s.setScrollY(s.scrollY - int(dy*scrollStep)) }

func (s *server) setScrollY(next int) {
	_, height := s.app.Size()
	bottom := float64(height)
	for _, b := range s.app.Boxes() {
		if b.Y+b.H > bottom {
			bottom = b.Y + b.H
		}
	}
	next = max(0, min(next, int(math.Ceil(bottom))-height))
	if next == s.scrollY {
		return
	}
	s.scrollY = next
	if observer, ok := s.app.(host.ScrollObserver); ok {
		observer.SetScrollOffset(0, next)
	}
	s.frameGen++
}

func (s *server) hover(w http.ResponseWriter, r *http.Request) {
	if !requireOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	x, okX := wheelFloat(r.PostFormValue("x"))
	y, okY := wheelFloat(r.PostFormValue("y"))
	if !okX || !okY {
		http.Error(w, "bad coordinates", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if x >= 0 && y >= 0 {
		y += float64(s.scrollY)
	}
	before := s.app.Generation()
	previousID := ""
	hoverer, hasHover := s.app.(host.Hoverer)
	if hasHover {
		previousID = hoverer.HoverID()
	}
	if err := s.app.Hover(r.Context(), x, y); err != nil {
		http.Error(w, "hover failed", http.StatusInternalServerError)
		return
	}
	if !hasHover || previousID != hoverer.HoverID() || before != s.app.Generation() {
		s.frameGen++
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) syncScroll() {
	if requester, ok := s.app.(host.ScrollRequester); ok {
		if move, pending := requester.TakeScroll(); pending {
			target := move.Y
			if !move.Absolute {
				target += s.scrollY
			}
			s.setScrollY(target)
			return
		}
	}
	if s.scrollY != 0 {
		s.setScrollY(s.scrollY)
	}
}
