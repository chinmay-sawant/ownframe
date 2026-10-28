package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// frameState is the JSON body of GET /state. It carries no boxes or field
// text, so the polling route exposes less than the opt-in debug route.
type frameState struct {
	Generation uint64 `json:"generation"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Ticking    bool   `json:"ticking"`
}

// state answers GET /state with a small snapshot. frameGen is the number of
// dirty ticks that changed retained operations without a page redraw, so it
// keeps the reported generation moving past the page's own count.
func (s *server) state(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reload(r.Context())
	state := s.frameSnapshotLocked()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Printf("write state: %s", sanitizeLine(err.Error()))
	}
}

// frameSnapshotLocked builds the state while the caller holds s.mu.
func (s *server) frameSnapshotLocked() frameState {
	front := host.SelectFront(s.app)

	return frameState{
		Generation: front.Generation + s.frameGen,
		Width:      front.Width,
		Height:     front.Height,
		Ticking:    s.ticking,
	}
}
