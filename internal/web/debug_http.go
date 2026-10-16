package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// debugState is the JSON body of GET /debug/state.
type debugState struct {
	Width      int            `json:"width"`
	Height     int            `json:"height"`
	Generation uint64         `json:"generation"`
	Fallback   bool           `json:"fallback"`
	Boxes      []layout.Box   `json:"boxes"`
	Stats      *host.Stats    `json:"stats,omitempty"`
	Ops        map[string]int `json:"ops,omitempty"`
	Runtime    map[string]any `json:"runtime,omitempty"`
}

// debug answers GET /debug/state: the page's size, generation, boxes, and
// stats, plus per-kind operation counts when a display list exists. It takes
// the server mutex, so it never reads the page mid-redraw.
//
// Panic recovery is installed at the server mux; keep this handler simple.
// Runtime sampling is guarded.
func (s *server) debug(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	front := page.SelectFront(s.app)
	state := debugState{
		Width:      front.Width,
		Height:     front.Height,
		Generation: front.Generation,
		Fallback:   front.Fallback,
		Boxes:      s.app.Boxes(),
	}

	if front.Display != nil {
		state.Ops = debugOpCounts(front.Display)
	}

	if insp, ok := s.app.(host.Inspector); ok {
		stats := insp.Stats()
		state.Stats = &stats
	}

	state.Runtime = runtimeSnapshot()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Printf("write debug state: %v", err)
	}
}
