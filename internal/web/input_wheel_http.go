package web

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// maxWheelValue bounds one wheel field before parsing.
const maxWheelValue = 64

// wheel forwards one wheel event in CSS pixels. A screen without a wheel
// handler accepts the request and ignores it.
func (s *server) wheel(w http.ResponseWriter, r *http.Request) {
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
	dx, okDX := wheelFloat(r.PostFormValue("dx"))
	dy, okDY := wheelFloat(r.PostFormValue("dy"))
	if !okX || !okY || !okDX || !okDY {
		http.Error(w, "bad wheel", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	handler, ok := s.app.(host.WheelHandler)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if _, err := handler.Wheel(r.Context(), x, y, dx, dy); err != nil {
		log.Printf("wheel: %s", sanitizeLine(err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// wheelFloat parses one finite wheel field.
func wheelFloat(raw string) (float64, bool) {
	if raw == "" || len(raw) > maxWheelValue {
		return 0, false
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false
	}

	return value, true
}
