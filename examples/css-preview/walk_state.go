package main

import (
	"fmt"
	"net/http"
)

// handleBox reports the box of one element. The boxes come from the
// snapshot the game loop published, so the route never reads the page
// while the loop swaps the boxes.
func (w *walker) handleBox(rw http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, b := range w.boxes {
		if b.ID == id {
			_, _ = fmt.Fprintf(rw, "%s %.1f %.1f %.1f %.1f\n", b.ID, b.X, b.Y, b.W, b.H)
			return
		}
	}
	_, _ = fmt.Fprintln(rw, "missing")
}

// handleState reports the counters the click script polls. The field order
// is a contract: scripts/css-preview-click-test.py reads the fields by
// position, so the order may not change.
func (w *walker) handleState(rw http.ResponseWriter, _ *http.Request) {
	w.mu.Lock()
	next, tr := w.next, w.trace
	w.mu.Unlock()
	_, _ = fmt.Fprintf(rw, "%d %d %s %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d\n",
		next, tr.started, tr.selected, tr.last, tr.redraws, tr.parses,
		tr.cascades, tr.layouts, tr.repaints, tr.ops, tr.boxes, tr.tmpl,
		tr.style, tr.display, tr.paint, tr.changed, tr.dirty, tr.alloc,
		tr.startedNS, tr.seq)
}
