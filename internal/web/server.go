// Package web shows the application HTML document at /.
// The PNG shell remains the fallback when the page has no executed source
// or the request sets dom=0. Requests deliver mouse, wheel, and keyboard
// events to a host.Screen. The native window is internal/window.
package web

import (
	"context"
	"html/template"
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type shellArea struct {
	Coords string
	Href   string
	Alt    string
	X      float64
	Y      float64
}

type shellData struct {
	Width      int
	Height     int
	Generation uint64
	Areas      []shellArea
	Reload     bool
}

type server struct {
	mu           sync.Mutex
	app          host.Screen
	shell        *template.Template
	lastNote     string
	perf         bool
	debugEnabled bool
	// ticking is true while the frame loop runs. frameGen advances when a
	// dirty tick changes the picture without a page redraw, so /state can
	// tell the shell to fetch a new frame.
	ticking                       bool
	frameGen                      uint64
	scrollY                       int
	viewportPNG                   []byte
	viewportGen                   uint64
	viewportWidth, viewportHeight int
	// limitMu guards the mutating-route throttle table and semaphore.
	limitMu      sync.Mutex
	throttleHits map[string]*throttleSeen
	inflight     chan struct{}
}

// Serve listens on addr and blocks. The page at / shows the application HTML
// document, or the PNG shell when the page has no executed source or the
// request sets dom=0. It forwards pointer, wheel, and key input to the
// screen. A screen that implements host.Ticker is ticked at about 60 Hz
// while Serve runs.
// Developer endpoints stay off; ServeWithOptions with Options.Perf turns
// on /debug/pprof/*.
func Serve(app host.Screen, addr string) error {
	return ServeWithOptions(app, addr, Options{})
}

// ServeWithContext stops the listener when ctx ends.
func ServeWithContext(ctx context.Context, app host.Screen, addr string) error {
	return ServeWithOptionsContext(ctx, app, addr, Options{})
}
