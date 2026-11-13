package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// events streams server-sent events to the browser. It is an additive
// invalidation hint: /state, /dom, /frame.png, and all POST routes keep
// their shapes. The live document refetches /dom only when domGeneration
// moves; the PNG shell refetches /frame.png only when generation moves.
// Idle servers send nothing but a 15s heartbeat comment.
func (s *server) events(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	if _, err := fmt.Fprint(w, "retry: 3000\n\n"); err != nil {
		return
	}
	fl.Flush()

	ch := s.subscribe(1)
	defer s.unsubscribe(ch)

	// Replay the current snapshot so a late joiner converges without
	// waiting one tick. This write happens in the handler goroutine with
	// no lock held.
	s.mu.Lock()
	replay, hasReplay := encodeEvent(s.frameSnapshotLocked())
	s.mu.Unlock()
	if hasReplay {
		if _, err := w.Write(replay); err != nil {
			return
		}
		fl.Flush()
	}

	hb := time.NewTicker(15 * time.Second)
	defer hb.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if _, err := w.Write(msg); err != nil {
				return
			}
			fl.Flush()
		case <-hb.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			fl.Flush()
		}
	}
}

// subscribe registers one /events stream. The channel is buffered size 1
// so a slow browser coalesces to the latest snapshot instead of stalling
// the tick loop.
func (s *server) subscribe(size int) chan []byte {
	if size < 1 {
		size = 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.subs == nil {
		s.subs = map[chan []byte]struct{}{}
	}
	ch := make(chan []byte, size)
	s.subs[ch] = struct{}{}
	return ch
}

// unsubscribe removes the stream and closes its channel.
func (s *server) unsubscribe(ch chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.subs == nil {
		return
	}
	delete(s.subs, ch)
	close(ch)
}

// broadcastLocked fans one encoded event out. The caller holds s.mu.
// Sends are non-blocking: a slow client drops to the latest snapshot and
// refetches authoritative /state on the next event it does receive.
func (s *server) broadcastLocked(msg []byte) {
	if len(s.subs) == 0 {
		return
	}
	for ch := range s.subs {
		select {
		case ch <- msg:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- msg:
			default:
			}
		}
	}
}

// publishIfChangedLocked compares the current snapshot to the last
// broadcast and emits one event on change. The caller holds s.mu.
// It never performs I/O, so the 60Hz idle tick stays a struct compare.
func (s *server) publishIfChangedLocked() {
	cur := s.frameSnapshotLocked()
	if cur == s.lastSent {
		return
	}
	msg, ok := encodeEvent(cur)
	if !ok {
		return
	}
	s.lastSent = cur
	s.broadcastLocked(msg)
}

// encodeEvent renders one SSE data frame for a snapshot.
func encodeEvent(state frameState) ([]byte, bool) {
	body, err := json.Marshal(state)
	if err != nil {
		return nil, false
	}
	msg := make([]byte, 0, len(body)+16)
	msg = append(msg, "data: "...)
	msg = append(msg, body...)
	msg = append(msg, '\n', '\n')
	return msg, true
}
