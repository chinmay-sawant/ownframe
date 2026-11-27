package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// The snapshot endpoint serves a picture of the desktop window as it is now.
// The window owns the page, so the HTTP handler never touches it: a request
// raises a flag and wakes the window, the frame hook answers on the goroutine
// that owns the page, and an idle window pays nothing while nobody asks.

const (
	snapshotWait          = 20 * time.Second
	snapshotPoll          = 10 * time.Millisecond
	snapshotHeaderTimeout = 5 * time.Second
)

var errSnapshotTimeout = errors.New("snapshot: the window did not answer")

type snapshotHub struct {
	mu   sync.Mutex
	want bool
	seq  uint64
	png  []byte
	wake func()
}

// request asks the window for a fresh picture and waits for it.
func (h *snapshotHub) request(ctx context.Context) ([]byte, error) {
	h.mu.Lock()
	before := h.seq
	h.want = true
	wake := h.wake
	h.mu.Unlock()
	if wake != nil {
		wake()
	}
	deadline := time.NewTimer(snapshotWait)
	defer deadline.Stop()
	tick := time.NewTicker(snapshotPoll)
	defer tick.Stop()
	for {
		h.mu.Lock()
		if h.seq != before {
			out := append([]byte(nil), h.png...)
			h.mu.Unlock()
			return out, nil
		}
		h.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("snapshot: %w", ctx.Err())
		case <-deadline.C:
			return nil, errSnapshotTimeout
		case <-tick.C:
		}
	}
}

// answer runs in the window's frame hook and publishes a picture when a
// request is waiting.
func (h *snapshotHub) answer(paint func() []byte) {
	h.mu.Lock()
	waiting := h.want
	h.want = false
	h.mu.Unlock()
	if !waiting {
		return
	}
	png := paint()
	h.mu.Lock()
	h.png = png
	h.seq++
	h.mu.Unlock()
}

const snapshotViewer = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>quarry: window snapshot</title>
<style>body{margin:0;background:#222;color:#ddd;font-family:sans-serif}button{margin:8px;padding:6px 14px}img{display:block}</style></head><body>
<button id="b" type="button">Refresh</button> <span>one picture per click or per GET /frame.png</span>
<img id="f" src="/frame.png" alt="window">
<script>document.getElementById("b").addEventListener("click",function(){document.getElementById("f").src="/frame.png?t="+Date.now();});</script>
</body></html>`

// serveSnapshots listens on a loopback address until ctx ends.
func serveSnapshots(ctx context.Context, addr string, hub *snapshotHub) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("snapshot address %q: %w", addr, err)
	}
	if ip := net.ParseIP(host); host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return fmt.Errorf("snapshot address %q: only loopback addresses are served", addr)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /frame.png", func(w http.ResponseWriter, r *http.Request) {
		png, err := hub.request(r.Context())
		if err != nil || len(png) == 0 {
			http.Error(w, "the window did not answer", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(png)
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(snapshotViewer))
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: snapshotHeaderTimeout}
	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("snapshot server: %w", err)
	}
	return nil
}
