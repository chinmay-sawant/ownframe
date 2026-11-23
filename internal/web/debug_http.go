package web

import (
	"net/http"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
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

// debug answers GET /debug/state under the server mutex. Reload errors
// are sanitized so file paths cannot leak verbatim to the client.
func (s *server) debug(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	s.mu.Lock()
	rp := s.debugReplyLocked()
	s.mu.Unlock()
	rp.send(w, "debug state")
}

func (s *server) debugReplyLocked() reply {
	front := host.SelectFront(s.app)
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
		stats.LastReloadError = sanitizeReload(stats.LastReloadError)
		state.Stats = &stats
	}
	if s.perf {
		state.Runtime = runtimeSnapshot()
	}
	return jsonReply("application/json", state, "debug state")
}
