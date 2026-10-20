package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// serve exposes the walker to the Python clicker. It listens only while
// CSS_PREVIEW_WALK is set and never runs in a normal example launch.
func (w *walker) serve() {
	http.HandleFunc("/goto", func(rw http.ResponseWriter, r *http.Request) {
		var i int
		fmt.Sscanf(r.URL.Query().Get("i"), "%d", &i)
		w.mu.Lock()
		w.next, w.scrollFor, w.aim = i, i, [3]int{-1, 0, 0}
		w.stopReq = true
		w.mu.Unlock()
		ebiten.ScheduleFrame()
	})
	http.HandleFunc("/aim", func(rw http.ResponseWriter, _ *http.Request) {
		w.mu.Lock()
		name := ""
		if w.aim[0] >= 0 && w.aim[0] < len(w.v.Catalog) {
			name = w.v.Catalog[w.aim[0]].Name
		}
		fmt.Fprintf(rw, "%d %d %d %s\n", w.aim[0], w.aim[1], w.aim[2], name)
		w.mu.Unlock()
	})
	http.HandleFunc("/cursor", func(rw http.ResponseWriter, _ *http.Request) {
		cx, cy := ebiten.CursorPosition()
		fmt.Fprintf(rw, "%d %d\n", cx, cy)
	})
	http.HandleFunc("/quit", func(rw http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(rw, "ok")
		go func() {
			time.Sleep(100 * time.Millisecond)
			os.Exit(0)
		}()
	})
	http.HandleFunc("/reset", func(rw http.ResponseWriter, _ *http.Request) {
		done := make(chan struct{})
		w.mu.Lock()
		w.resetReq, w.resetDone = true, done
		w.mu.Unlock()
		<-done
		fmt.Fprintln(rw, "ok")
	})
	http.HandleFunc("/state", func(rw http.ResponseWriter, _ *http.Request) {
		w.mu.Lock()
		next, tr := w.next, w.trace
		w.mu.Unlock()
		fmt.Fprintf(rw, "%d %d %s %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d\n",
			next, tr.started, tr.selected, tr.last, tr.redraws, tr.parses,
			tr.cascades, tr.layouts, tr.repaints, tr.ops, tr.boxes, tr.tmpl,
			tr.style, tr.display, tr.paint, tr.changed, tr.dirty, tr.alloc,
			tr.startedNS, tr.seq)
	})
	_ = http.ListenAndServe("127.0.0.1:6971", nil)
}
