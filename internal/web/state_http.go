package web

import (
	"encoding/json"
	"log"
	"net/http"
)

// frameState is the JSON body of GET /state. It carries no boxes or field
// text, so the polling route exposes less than the opt-in debug route.
type frameState struct {
	Generation uint64 `json:"generation"`
	// DomGeneration is the page generation without frameGen. The live
	// document only changes on a real Redraw, so the boot script watches
	// this field for /dom and ignores retained-op frameGen bumps that only
	// invalidate the PNG frame.
	DomGeneration uint64 `json:"domGeneration"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Ticking       bool   `json:"ticking"`
	// AudioGen is the published clip generation, or 0 when the screen has
	// no Audio method or no clip.
	AudioGen uint64 `json:"audioGen"`
	// AudioPlaying is whether the preview should be playing that clip.
	AudioPlaying bool `json:"audioPlaying"`
	// AudioVolume is the preview loudness from 0 to 1. It is 1 until the
	// page publishes a level.
	AudioVolume float64 `json:"audioVolume"`
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
	width, height := s.app.Size()
	playing, volume := playbackState(s.app)

	return frameState{
		Generation:    s.app.Generation() + s.frameGen,
		DomGeneration: s.app.Generation(),
		Width:         width,
		Height:        height,
		Ticking:       s.ticking,
		AudioGen:      clipGeneration(s.app),
		AudioPlaying:  playing,
		AudioVolume:   volume,
	}
}

func playbackState(app any) (bool, float64) {
	src, ok := app.(interface {
		AudioPlayback() (playing bool, volume float64)
	})
	if !ok {
		return false, 1
	}

	return src.AudioPlayback()
}
