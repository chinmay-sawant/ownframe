package main

import (
	"fmt"
	"net/http"

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
		w.mu.Unlock()
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
	http.HandleFunc("/state", func(rw http.ResponseWriter, _ *http.Request) {
		w.mu.Lock()
		started := 0
		for _, row := range w.v.Catalog {
			if w.v.Started[row.Button] {
				started++
			}
		}
		fmt.Fprintf(rw, "%d %d %s %d\n", w.next, started, w.v.Selected,
			w.p.Stats().LastRedraw.Microseconds())
		w.mu.Unlock()
	})
	_ = http.ListenAndServe("127.0.0.1:6971", nil)
}
