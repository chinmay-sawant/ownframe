package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// The snapshot endpoint serves a picture of whatever the desktop window shows
// right now: the open tab, hover, focus, the carets and every live patch. The
// window owns the page, so the HTTP handler never touches it. It asks the
// window's own frame loop for a picture and waits. The loop answers from a
// live session, on the goroutine that owns the page, only while someone is
// waiting, so an idle window pays nothing: the session never asks the window
// for a frame, and the page viewer does not reload on its own. A request wakes
// the window once.

const (
	// snapshotWait is how long a request waits for the window before it gives up.
	// A busy page can take seconds to paint, so the wait is generous.
	snapshotWait = 20 * time.Second
	// snapshotPoll is how often a waiting request checks whether the window answered.
	snapshotPoll = 10 * time.Millisecond
	// snapshotHeaderTimeout bounds how long a client may take to send request headers.
	snapshotHeaderTimeout = 5 * time.Second
)

// errSnapshotTimeout means the window loop did not answer in time.
var errSnapshotTimeout = errors.New("snapshot: the window did not answer")

// snapshotHub carries one request at a time from the HTTP side to the window.
type snapshotHub struct {
	mu   sync.Mutex
	want bool
	seq  uint64
	png  []byte
	// wake asks the window for one frame so a waiting request is served
	// promptly. It may be nil.
	wake func()
}

func newSnapshotHub() *snapshotHub { return &snapshotHub{} }

// Request asks the window for a fresh picture and waits for it.
func (h *snapshotHub) Request(ctx context.Context) ([]byte, error) {
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

// take reports whether a request is waiting and clears it.
func (h *snapshotHub) take() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	waiting := h.want
	h.want = false

	return waiting
}

// publish stores the picture a request is waiting for.
func (h *snapshotHub) publish(png []byte) {
	h.mu.Lock()
	h.png = png
	h.seq++
	h.mu.Unlock()
}

// snapshotSession answers the hub from inside the window's frame loop.
type snapshotSession struct{ hub *snapshotHub }

// TabID is empty: the session works on every tab.
func (s *snapshotSession) TabID() string { return "" }

// Frame takes a picture when a request is waiting. The probe paints the
// retained display list, so live patches and hover are in it. It returns the
// zero time: the session never asks for a frame of its own.
func (s *snapshotSession) Frame(
	ctx context.Context, page *ownframe.Page, _ liveInput, _ time.Time,
) (time.Time, error) {
	if s.hub.take() {
		s.hub.publish(page.Probe().PNGWithContext(ctx))
	}

	return time.Time{}, nil
}

// snapshotViewer shows the picture once. It never reloads by itself: the
// button asks for a new one, so nothing happens until someone asks.
const snapshotViewer = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>blinkless all: window snapshot</title>
<style>body{margin:0;background:#222;color:#ddd;font-family:sans-serif}
button{margin:8px;padding:6px 14px;font-size:14px}img{display:block}</style></head><body>
<button id="b" type="button">Refresh</button> <span>one picture per click or per GET /frame.png, no automatic reload</span>
<img id="f" src="/frame.png" alt="window">
<script>document.getElementById("b").addEventListener("click",function(){document.getElementById("f").src="/frame.png?t="+Date.now();});</script>
</body></html>`

// handlerPNG serves the window's current picture.
func handlerPNG(hub *snapshotHub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		png, err := hub.Request(r.Context())
		if err != nil || len(png) == 0 {
			http.Error(w, "the window did not answer", http.StatusServiceUnavailable)

			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(png)
	})
}

// serveSnapshots listens on addr, loopback only, until ctx ends.
func serveSnapshots(ctx context.Context, addr string, hub *snapshotHub) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("snapshot address %q: %w", addr, err)
	}

	if ip := net.ParseIP(host); host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return fmt.Errorf("snapshot address %q: only loopback addresses are served", addr) //nolint:err113 // one-off
	}

	mux := http.NewServeMux()

	mux.Handle("GET /frame.png", handlerPNG(hub))

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

	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("snapshot server: %w", err)
	}

	return nil
}
