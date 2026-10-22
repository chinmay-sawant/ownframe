package main

import (
	"net/http"
	"time"
)

// walkAddr is the loopback address the click-script walker serves on.
const walkAddr = "127.0.0.1:6971"

// walkTimeouts bounds the walker server I/O; the click script polls it.
const (
	walkReadHeaderTimeout = 3 * time.Second
	walkReadTimeout       = 5 * time.Second
	walkWriteTimeout      = 5 * time.Second
	// walkQuitDelay lets the quit reply flush before the process exits.
	walkQuitDelay = 100 * time.Millisecond
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
	srv := &http.Server{Addr: walkAddr, Handler: mux,
		ReadHeaderTimeout: walkReadHeaderTimeout, ReadTimeout: walkReadTimeout, WriteTimeout: walkWriteTimeout}
	_ = srv.ListenAndServe()
}
