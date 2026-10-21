package main

import (
	"net/http"
)

// serve exposes the walker to the Python clicker. It listens only while
// CSS_PREVIEW_WALK is set and never runs in a normal example launch. The
// routes live on their own mux, so the walker never touches the default
// one that another package may register on.
func (w *walker) serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/goto", w.handleGoto)
	mux.HandleFunc("/aim", w.handleAim)
	mux.HandleFunc("/cursor", w.handleCursor)
	mux.HandleFunc("/quit", w.handleQuit)
	mux.HandleFunc("/reset", w.handleReset)
	mux.HandleFunc("/box", w.handleBox)
	mux.HandleFunc("/state", w.handleState)
	_ = http.ListenAndServe("127.0.0.1:6971", mux)
}
