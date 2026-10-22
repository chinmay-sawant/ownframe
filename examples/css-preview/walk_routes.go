package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// handleGoto points the walker at a catalog row and stops the running
// sample, so the next measurement is one click.
func (w *walker) handleGoto(_ http.ResponseWriter, r *http.Request) {
	var i int
	_, _ = fmt.Sscanf(r.URL.Query().Get("i"), "%d", &i)
	w.mu.Lock()
	w.next, w.scrollFor, w.aim = i, i, [3]int{-1, 0, 0}
	w.stopReq = true
	w.mu.Unlock()
	ebiten.ScheduleFrame() //nolint:staticcheck // reason: the walker needs one prompt frame after a goto; SetScreenClearedEveryFrame(false) only skips clearing and cannot wake event-wait.
}

// handleAim reports the client point of the row the walker aims at, with
// the property name the click script matches on.
func (w *walker) handleAim(rw http.ResponseWriter, _ *http.Request) {
	w.mu.Lock()
	defer w.mu.Unlock()
	name := ""
	if w.aim[0] >= 0 && w.aim[0] < len(w.v.Catalog) {
		name = w.v.Catalog[w.aim[0]].Name
	}
	_, _ = fmt.Fprintf(rw, "%d %d %d %s\n", w.aim[0], w.aim[1], w.aim[2], name)
}

// handleCursor reports the client cursor position, which the click script
// uses to calibrate the window origin.
func (w *walker) handleCursor(rw http.ResponseWriter, _ *http.Request) {
	cx, cy := ebiten.CursorPosition()
	_, _ = fmt.Fprintf(rw, "%d %d\n", cx, cy)
}

// handleQuit closes the app after the script's last pass.
func (w *walker) handleQuit(rw http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintln(rw, "ok")
	go func() {
		time.Sleep(walkQuitDelay)
		os.Exit(0)
	}()
}

// handleReset clears the walk state. The game loop does the work, so the
// route waits for it before it answers.
func (w *walker) handleReset(rw http.ResponseWriter, _ *http.Request) {
	done := make(chan struct{})
	w.mu.Lock()
	w.resetReq, w.resetDone = true, done
	w.mu.Unlock()
	<-done
	_, _ = fmt.Fprintln(rw, "ok")
}
